package assetcontroller

import (
	"context"
	"fmt"
	apicommon "helloworld/api/common"
	"strings"
	"time"

	apiasset "helloworld/api/asset"
	"helloworld/pkg/dao/assetdao"
	assetmodel "helloworld/pkg/model/asset"
	"mall/kit/auth"
	"mall/kit/codeerror"
	"mall/kit/logger"
	"mall/user/dao"
)

var (
	_ apiasset.AssetHTTPServer      = (*Controller)(nil)
	_ apiasset.AdminAssetHTTPServer = (*Controller)(nil)
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) GetUserAssets(ctx context.Context, req *apiasset.GetUserAssetsRequest) (*apiasset.GetUserAssetsReply, error) {
	info := auth.MustFromContext(ctx)
	assetType := req.Type
	if assetType < 0 {
		logger.Errorf("GetUserAssets invalid type=%d userID=%d", assetType, info.UserID)
		return nil, codeerror.ErrorParam
	}

	list, err := assetdao.NewAssetDao(nil).ListByUserID(info.UserID, assetType)
	if err != nil {
		logger.Errorf("GetUserAssets ListByUserID userID=%d type=%d err=%v", info.UserID, assetType, err)
		return nil, err
	}

	if (assetType == 0 || assetType == apicommon.AssetType_ASSET_TYPE_POINT) && !hasType(list, apicommon.AssetType_ASSET_TYPE_POINT) {
		list = append(list, assetmodel.UserAsset{
			UserID:  info.UserID,
			Type:    apicommon.AssetType_ASSET_TYPE_POINT,
			Balance: 0,
		})
	}

	assets := make([]*apiasset.AssetInfo, 0, len(list))
	for i := range list {
		assets = append(assets, toAssetInfo(&list[i]))
	}
	return &apiasset.GetUserAssetsReply{Assets: assets}, nil
}

func (c *Controller) GetAssetHistory(ctx context.Context, req *apiasset.GetAssetHistoryRequest) (*apiasset.GetAssetHistoryReply, error) {
	info := auth.MustFromContext(ctx)
	assetType := req.Type
	if assetType < 0 {
		logger.Errorf("GetAssetHistory invalid type=%d userID=%d", assetType, info.UserID)
		return nil, codeerror.ErrorParam
	}

	list, total, err := assetdao.NewHistoryDao(nil).List(info.UserID, assetType, int(req.Page), int(req.PageSize))
	if err != nil {
		logger.Errorf("GetAssetHistory List userID=%d type=%d err=%v", info.UserID, assetType, err)
		return nil, err
	}

	out := make([]*apiasset.AssetHistory, 0, len(list))
	for i := range list {
		out = append(out, toHistory(&list[i]))
	}
	return &apiasset.GetAssetHistoryReply{
		List:  out,
		Total: total,
	}, nil
}

func (c *Controller) AdminGetUserAssets(ctx context.Context, req *apiasset.AdminGetUserAssetsRequest) (*apiasset.GetUserAssetsReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminGetUserAssets RequireAdmin err=%v", err)
		return nil, err
	}
	if req.UserId <= 0 {
		logger.Errorf("AdminGetUserAssets invalid user_id=%d", req.UserId)
		return nil, codeerror.ErrorParam
	}
	if _, err := userdao.NewUserDao(nil).GetByID(req.UserId); err != nil {
		logger.Errorf("AdminGetUserAssets GetByID userID=%d err=%v", req.UserId, err)
		return nil, err
	}

	assetType := req.Type
	if assetType < 0 {
		logger.Errorf("AdminGetUserAssets invalid type=%d userID=%d", assetType, req.UserId)
		return nil, codeerror.ErrorParam
	}

	list, err := assetdao.NewAssetDao(nil).ListByUserID(req.UserId, assetType)
	if err != nil {
		logger.Errorf("AdminGetUserAssets ListByUserID userID=%d type=%d err=%v", req.UserId, assetType, err)
		return nil, err
	}
	if (assetType == 0 || assetType == apicommon.AssetType_ASSET_TYPE_POINT) && !hasType(list, apicommon.AssetType_ASSET_TYPE_POINT) {
		list = append(list, assetmodel.UserAsset{
			UserID:  req.UserId,
			Type:    apicommon.AssetType_ASSET_TYPE_POINT,
			Balance: 0,
		})
	}

	assets := make([]*apiasset.AssetInfo, 0, len(list))
	for i := range list {
		assets = append(assets, toAssetInfo(&list[i]))
	}
	return &apiasset.GetUserAssetsReply{Assets: assets}, nil
}

func (c *Controller) AdminGetAssetHistory(ctx context.Context, req *apiasset.AdminGetAssetHistoryRequest) (*apiasset.GetAssetHistoryReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminGetAssetHistory RequireAdmin err=%v", err)
		return nil, err
	}
	if req.UserId <= 0 {
		logger.Errorf("AdminGetAssetHistory invalid user_id=%d", req.UserId)
		return nil, codeerror.ErrorParam
	}
	if _, err := userdao.NewUserDao(nil).GetByID(req.UserId); err != nil {
		logger.Errorf("AdminGetAssetHistory GetByID userID=%d err=%v", req.UserId, err)
		return nil, err
	}

	assetType := req.Type
	if assetType < 0 {
		logger.Errorf("AdminGetAssetHistory invalid type=%d userID=%d", assetType, req.UserId)
		return nil, codeerror.ErrorParam
	}

	list, total, err := assetdao.NewHistoryDao(nil).List(req.UserId, assetType, int(req.Page), int(req.PageSize))
	if err != nil {
		logger.Errorf("AdminGetAssetHistory List userID=%d type=%d err=%v", req.UserId, assetType, err)
		return nil, err
	}

	out := make([]*apiasset.AssetHistory, 0, len(list))
	for i := range list {
		out = append(out, toHistory(&list[i]))
	}
	return &apiasset.GetAssetHistoryReply{
		List:  out,
		Total: total,
	}, nil
}

func (c *Controller) AdminGrantPoints(ctx context.Context, req *apiasset.AdminGrantPointsRequest) (*apiasset.AdminGrantPointsReply, error) {
	admin, err := auth.RequireAdmin(ctx)
	if err != nil {
		logger.Errorf("AdminGrantPoints RequireAdmin err=%v", err)
		return nil, err
	}
	if req.UserId <= 0 || req.Points <= 0 {
		logger.Errorf("AdminGrantPoints invalid user_id=%d points=%d adminID=%d", req.UserId, req.Points, admin.UserID)
		return nil, codeerror.ErrorParam
	}
	if _, err := userdao.NewUserDao(nil).GetByID(req.UserId); err != nil {
		logger.Errorf("AdminGrantPoints GetByID userID=%d err=%v", req.UserId, err)
		return nil, err
	}

	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		remark = "后台手动加积分"
	}
	idempotentID := fmt.Sprintf("admin_grant_points_%d_%d_%d", admin.UserID, req.UserId, time.Now().UnixNano())
	asset, err := assetdao.Change(assetdao.ChangeReq{
		UserID:       req.UserId,
		Type:         apicommon.AssetType_ASSET_TYPE_POINT,
		Change:       req.Points,
		IdempotentID: idempotentID,
		BizType:      "admin_grant_points",
		BizID:        fmt.Sprintf("%d", admin.UserID),
		Remark:       remark,
	})
	if err != nil {
		logger.Errorf("AdminGrantPoints Change userID=%d points=%d adminID=%d err=%v", req.UserId, req.Points, admin.UserID, err)
		return nil, err
	}
	return &apiasset.AdminGrantPointsReply{Asset: toAssetInfo(asset)}, nil
}

func hasType(list []assetmodel.UserAsset, t apicommon.AssetType) bool {
	for i := range list {
		if list[i].Type == t {
			return true
		}
	}
	return false
}

func toAssetInfo(a *assetmodel.UserAsset) *apiasset.AssetInfo {
	if a == nil {
		return nil
	}
	return &apiasset.AssetInfo{
		UserId:  a.UserID,
		Type:    a.Type,
		Balance: a.Balance,
		Ct:      a.Ct,
		Ut:      a.Ut,
	}
}

func toHistory(h *assetmodel.UserAssetHistory) *apiasset.AssetHistory {
	if h == nil {
		return nil
	}
	return &apiasset.AssetHistory{
		Id:           h.ID,
		UserId:       h.UserID,
		Type:         h.Type,
		Change:       h.Change,
		Balance:      h.Balance,
		IdempotentId: h.IdempotentID,
		BizType:      h.BizType,
		BizId:        h.BizID,
		Remark:       h.Remark,
		Ct:           h.Ct,
	}
}
