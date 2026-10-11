package usercontroller

import (
	"context"
	"strings"

	apiuser "helloworld/api/user"
	"helloworld/base/codeerror"
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
		return nil, codeerror.ErrorParam
	}
	if nickname == "" {
		nickname = username
	}
	u, err := userdao.CreateUserWithPassword(nickname, username, password)
	if err != nil {
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
		return nil, codeerror.ErrorParam
	}
	login, err := userdao.GetLoginByPasswordAccount(username)
	if err != nil {
		return nil, err
	}
	if login.Credential != password {
		return nil, codeerror.ErrorPassword
	}
	u, err := userdao.GetByID(login.UserID)
	if err != nil {
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
		return nil, err
	}
	username, err := resolveUsername(u.ID, info.Username)
	if err != nil {
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
		return nil, err
	}
	username, err := resolveUsername(u.ID, info.Username)
	if err != nil {
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
		return fallback, nil
	}
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
