package usercontroller

import (
	"context"
	"strings"

	"mall/kit/auth"
	"mall/kit/codeerror"
	"mall/kit/logger"
	apiuser "mall/user/api/user"
	"mall/user/dao"
	usermodel "mall/user/model"
)

var (
	_ apiuser.UserHTTPServer      = (*Controller)(nil)
	_ apiuser.AdminUserHTTPServer = (*Controller)(nil)
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) Register(ctx context.Context, req *apiuser.RegisterRequest) (*apiuser.RegisterReply, error) {
	username := strings.TrimSpace(req.Username)
	password := req.Password
	nickname := strings.TrimSpace(req.Nickname)
	if username == "" || password == "" {
		logger.Errorf("Register param error username=%q password_empty=%v", username, password == "")
		return nil, codeerror.ErrorParam
	}
	if nickname == "" {
		nickname = username
	}
	u, err := userdao.CreateUserWithPassword(nickname, username, password)
	if err != nil {
		logger.Errorf("Register CreateUserWithPassword username=%s err=%v", username, err)
		return nil, err
	}
	return &apiuser.RegisterReply{
		User: toUserInfo(u, username),
	}, nil
}

func (c *Controller) Login(ctx context.Context, req *apiuser.LoginRequest) (*apiuser.LoginReply, error) {
	username := strings.TrimSpace(req.Username)
	password := req.Password
	if username == "" || password == "" {
		logger.Errorf("Login param error username=%q password_empty=%v", username, password == "")
		return nil, codeerror.ErrorParam
	}
	login, err := userdao.NewLoginDao(nil).GetByPasswordAccount(username)
	if err != nil {
		logger.Errorf("Login GetByPasswordAccount username=%s err=%v", username, err)
		return nil, err
	}
	if login.Credential != password {
		logger.Errorf("Login password mismatch username=%s userID=%d", username, login.UserID)
		return nil, codeerror.ErrorPassword
	}
	u, err := userdao.NewUserDao(nil).GetByID(login.UserID)
	if err != nil {
		logger.Errorf("Login GetByID userID=%d err=%v", login.UserID, err)
		return nil, err
	}
	if u.Banned() {
		logger.Errorf("Login user banned userID=%d banUntil=%d", u.ID, u.BanUntil)
		return nil, codeerror.ErrorUserBanned
	}
	token, err := auth.GenerateToken(u.ID, username, u.Admin())
	if err != nil {
		return nil, codeerror.OrSystem(err, "Login GenerateToken userID=%d", u.ID)
	}
	return &apiuser.LoginReply{
		User:        toUserInfo(u, username),
		AccessToken: token,
	}, nil
}

// RefreshToken 用当前仍有效的登录态换发新 token。身份以库中的用户和管理员标记为准。
func (c *Controller) RefreshToken(ctx context.Context, _ *apiuser.RefreshTokenRequest) (*apiuser.RefreshTokenReply, error) {
	info := auth.MustFromContext(ctx)
	u, err := userdao.NewUserDao(nil).GetByID(info.UserID)
	if err != nil {
		logger.Errorf("RefreshToken GetByID userID=%d err=%v", info.UserID, err)
		return nil, err
	}
	if u.Banned() {
		logger.Errorf("RefreshToken user banned userID=%d banUntil=%d", u.ID, u.BanUntil)
		return nil, codeerror.ErrorUserBanned
	}
	login, err := userdao.NewLoginDao(nil).GetPasswordByUserID(u.ID)
	if err != nil {
		logger.Errorf("RefreshToken GetPasswordByUserID userID=%d err=%v", u.ID, err)
		return nil, err
	}
	token, err := auth.GenerateToken(u.ID, login.Identifier, u.Admin())
	if err != nil {
		return nil, codeerror.OrSystem(err, "RefreshToken GenerateToken userID=%d", u.ID)
	}
	return &apiuser.RefreshTokenReply{
		AccessToken: token,
	}, nil
}

func (c *Controller) GetUserInfo(ctx context.Context, req *apiuser.GetUserInfoRequest) (*apiuser.GetUserInfoReply, error) {
	info := auth.MustFromContext(ctx)
	uid := req.Id
	if uid <= 0 {
		uid = info.UserID
	}
	u, err := userdao.NewUserDao(nil).GetByID(uid)
	if err != nil {
		logger.Errorf("GetUserInfo GetByID uid=%d err=%v", uid, err)
		return nil, err
	}
	username, err := resolveUsername(u.ID, info.Username)
	if err != nil {
		logger.Errorf("GetUserInfo resolveUsername userID=%d err=%v", u.ID, err)
		return nil, err
	}
	return &apiuser.GetUserInfoReply{
		User: toUserInfo(u, username),
	}, nil
}

func (c *Controller) UpdateUserInfo(ctx context.Context, req *apiuser.UpdateUserInfoRequest) (*apiuser.UpdateUserInfoReply, error) {
	info := auth.MustFromContext(ctx)
	nickname := strings.TrimSpace(req.Nickname)
	avatar := avatarIn(req.Avatar)
	u, err := userdao.NewUserDao(nil).Update(info.UserID, nickname, req.Sex, avatar)
	if err != nil {
		logger.Errorf("UpdateUserInfo Update userID=%d err=%v", info.UserID, err)
		return nil, err
	}
	username, err := resolveUsername(u.ID, info.Username)
	if err != nil {
		logger.Errorf("UpdateUserInfo resolveUsername userID=%d err=%v", u.ID, err)
		return nil, err
	}
	return &apiuser.UpdateUserInfoReply{
		User: toUserInfo(u, username),
	}, nil
}

func (c *Controller) ChangePassword(ctx context.Context, req *apiuser.ChangePasswordRequest) (*apiuser.ChangePasswordReply, error) {
	info := auth.MustFromContext(ctx)
	oldPassword := req.GetOldPassword()
	newPassword := req.GetNewPassword()
	if oldPassword == "" || newPassword == "" {
		logger.Errorf("ChangePassword empty password userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	if oldPassword == newPassword {
		logger.Errorf("ChangePassword same password userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	login, err := userdao.NewLoginDao(nil).GetPasswordByUserID(info.UserID)
	if err != nil {
		logger.Errorf("ChangePassword GetPasswordByUserID userID=%d err=%v", info.UserID, err)
		return nil, err
	}
	if login.Credential != oldPassword {
		logger.Errorf("ChangePassword old password mismatch userID=%d", info.UserID)
		return nil, codeerror.ErrorPassword
	}
	if err := userdao.NewLoginDao(nil).UpdatePasswordCredential(info.UserID, newPassword); err != nil {
		logger.Errorf("ChangePassword UpdatePasswordCredential userID=%d err=%v", info.UserID, err)
		return nil, err
	}
	return &apiuser.ChangePasswordReply{}, nil
}

func resolveUsername(userID int64, fallback string) (string, error) {
	login, err := userdao.NewLoginDao(nil).GetPasswordByUserID(userID)
	if err == nil {
		return login.Identifier, nil
	}
	if ce, ok := codeerror.As(err); ok && ce.Is(codeerror.ErrorUserNotFound) {
		logger.Errorf("resolveUsername login not found userID=%d, use fallback", userID)
		return fallback, nil
	}
	logger.Errorf("resolveUsername userID=%d err=%v", userID, err)
	return "", err
}

func toUserInfo(u *usermodel.User, username string) *apiuser.UserInfo {
	if u == nil {
		return nil
	}
	return &apiuser.UserInfo{
		Id:       u.ID,
		Username: username,
		Nickname: u.Nickname,
		Sex:      u.Sex,
		Ct:       u.Ct,
		Ut:       u.Ut,
		IsAdmin:  u.Admin(),
		BanUntil: u.BanUntil,
		Banned:   u.Banned(),
		Avatar:   avatarOut(u.Avatar),
	}
}
