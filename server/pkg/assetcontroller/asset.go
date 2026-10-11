package assetcontroller

import (
	"context"

	apiasset "helloworld/api/asset"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/pkg/dao/assetdao"
	assetmodel "helloworld/pkg/model/asset"
)

var _ apiasset.AssetHTTPServer = (*Controller)(nil)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) GetUserAssets(ctx context.Context, req *apiasset.GetUserAssetsRequest) (*apiasset.GetUserAssetsReply, error) {
	info := auth.MustFromContext(ctx)
	assetType := int32(req.Type)
	if assetType < 0 {
		logger.Errorf("GetUserAssets invalid type=%d userID=%d", assetType, info.UserID)
		return nil, codeerror.ErrorParam
	}

	list, err := assetdao.NewAssetDao(nil).ListByUserID(info.UserID, assetType)
	if err != nil {
		logger.Errorf("GetUserAssets ListByUserID userID=%d type=%d err=%v", info.UserID, assetType, err)
		return nil, err
	}

	if (assetType == 0 || assetType == assetmodel.TypePoint) && !hasType(list, assetmodel.TypePoint) {
		list = append(list, assetmodel.UserAsset{
			UserID:  info.UserID,
			Type:    assetmodel.TypePoint,
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
	assetType := int32(req.Type)
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

func hasType(list []assetmodel.UserAsset, t int32) bool {
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
		Type:    apiasset.AssetType(a.Type),
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
		Type:         apiasset.AssetType(h.Type),
		Change:       h.Change,
		Balance:      h.Balance,
		IdempotentId: h.IdempotentID,
		BizType:      h.BizType,
		BizId:        h.BizID,
		Remark:       h.Remark,
		Ct:           h.Ct,
	}
}
