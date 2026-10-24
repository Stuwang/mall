package usercontroller

import (
	"context"
	"strings"
	"time"

	apiuser "helloworld/api/user"
	"helloworld/pkg/dao/userdao"
	usermodel "helloworld/pkg/model/user"
	"mall/kit/auth"
	"mall/kit/codeerror"
	"mall/kit/logger"
)

func (c *Controller) AdminListUsers(ctx context.Context, req *apiuser.AdminListUsersRequest) (*apiuser.AdminListUsersReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminListUsers RequireAdmin err=%v", err)
		return nil, err
	}
	list, total, err := userdao.NewUserDao(nil).List(userdao.UserQuery{
		UserID:   req.UserId,
		Keyword:  strings.TrimSpace(req.Keyword),
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		logger.Errorf("AdminListUsers err=%v", err)
		return nil, err
	}
	infos, err := toUserInfos(list)
	if err != nil {
		logger.Errorf("AdminListUsers toUserInfos err=%v", err)
		return nil, err
	}
	return &apiuser.AdminListUsersReply{List: infos, Total: total}, nil
}

func (c *Controller) AdminGetUser(ctx context.Context, req *apiuser.AdminGetUserRequest) (*apiuser.GetUserInfoReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminGetUser RequireAdmin err=%v", err)
		return nil, err
	}
	if req.UserId <= 0 {
		logger.Errorf("AdminGetUser invalid user_id=%d", req.UserId)
		return nil, codeerror.ErrorParam
	}
	u, err := userdao.NewUserDao(nil).GetByID(req.UserId)
	if err != nil {
		logger.Errorf("AdminGetUser GetByID userID=%d err=%v", req.UserId, err)
		return nil, err
	}
	username, err := resolveUsername(u.ID, "")
	if err != nil {
		logger.Errorf("AdminGetUser resolveUsername userID=%d err=%v", u.ID, err)
		return nil, err
	}
	return &apiuser.GetUserInfoReply{User: toUserInfo(u, username)}, nil
}

func (c *Controller) AdminBanUser(ctx context.Context, req *apiuser.AdminBanUserRequest) (*apiuser.GetUserInfoReply, error) {
	admin, err := auth.RequireAdmin(ctx)
	if err != nil {
		logger.Errorf("AdminBanUser RequireAdmin err=%v", err)
		return nil, err
	}
	if req.UserId <= 0 {
		logger.Errorf("AdminBanUser invalid user_id=%d", req.UserId)
		return nil, codeerror.ErrorParam
	}
	if req.UserId == admin.UserID {
		logger.Errorf("AdminBanUser cannot ban self userID=%d", admin.UserID)
		return nil, codeerror.ErrorParam.Errorf("cannot ban self")
	}
	if req.BanUntil < 0 {
		logger.Errorf("AdminBanUser invalid ban_until=%d userID=%d", req.BanUntil, req.UserId)
		return nil, codeerror.ErrorParam
	}
	if req.BanUntil > 0 && req.BanUntil <= time.Now().Unix() {
		logger.Errorf("AdminBanUser ban_until not in future banUntil=%d userID=%d", req.BanUntil, req.UserId)
		return nil, codeerror.ErrorParam.Errorf("ban_until must be in the future")
	}
	u, err := userdao.NewUserDao(nil).SetBanUntil(req.UserId, req.BanUntil)
	if err != nil {
		logger.Errorf("AdminBanUser SetBanUntil userID=%d banUntil=%d err=%v", req.UserId, req.BanUntil, err)
		return nil, err
	}
	username, err := resolveUsername(u.ID, "")
	if err != nil {
		logger.Errorf("AdminBanUser resolveUsername userID=%d err=%v", u.ID, err)
		return nil, err
	}
	return &apiuser.GetUserInfoReply{User: toUserInfo(u, username)}, nil
}

func toUserInfos(list []usermodel.User) ([]*apiuser.UserInfo, error) {
	if len(list) == 0 {
		return []*apiuser.UserInfo{}, nil
	}
	ids := make([]int64, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].ID)
	}
	nameMap, err := userdao.NewLoginDao(nil).MapPasswordUsernameByUserIDs(ids)
	if err != nil {
		return nil, err
	}
	out := make([]*apiuser.UserInfo, 0, len(list))
	for i := range list {
		out = append(out, toUserInfo(&list[i], nameMap[list[i].ID]))
	}
	return out, nil
}
