package usercontroller

import (
	"context"
	"strings"

	apiuser "helloworld/api/user"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/pkg/dao/userdao"
	usermodel "helloworld/pkg/model/user"
)

var _ apiuser.UserHTTPServer = (*Controller)(nil)

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
	login, err := userdao.GetLoginByPasswordAccount(username)
	if err != nil {
		logger.Errorf("Login GetLoginByPasswordAccount username=%s err=%v", username, err)
		return nil, err
	}
	if login.Credential != password {
		logger.Errorf("Login password mismatch username=%s userID=%d", username, login.UserID)
		return nil, codeerror.ErrorPassword
	}
	u, err := userdao.GetByID(login.UserID)
	if err != nil {
		logger.Errorf("Login GetByID userID=%d err=%v", login.UserID, err)
		return nil, err
	}
	token, err := auth.GenerateToken(u.ID, username)
	if err != nil {
		return nil, codeerror.OrSystem(err, "Login GenerateToken userID=%d", u.ID)
	}
	return &apiuser.LoginReply{
		User:        toUserInfo(u, username),
		AccessToken: token,
	}, nil
}

func (c *Controller) GetUserInfo(ctx context.Context, req *apiuser.GetUserInfoRequest) (*apiuser.GetUserInfoReply, error) {
	info := auth.MustFromContext(ctx)
	uid := req.Id
	if uid <= 0 {
		uid = info.UserID
	}
	u, err := userdao.GetByID(uid)
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
	u, err := userdao.UpdateUser(info.UserID, strings.TrimSpace(req.Nickname), req.Sex)
	if err != nil {
		logger.Errorf("UpdateUserInfo UpdateUser userID=%d err=%v", info.UserID, err)
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

func resolveUsername(userID int64, fallback string) (string, error) {
	login, err := userdao.GetPasswordLoginByUserID(userID)
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
	}
}
