package misccontroller

import (
	"context"

	apimisc "helloworld/api/misc"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/pkg/dao/miscdao"
	"helloworld/pkg/mediastore"
	miscmodel "helloworld/pkg/model/misc"
)

var (
	_ apimisc.MiscHTTPServer      = (*Controller)(nil)
	_ apimisc.AdminMiscHTTPServer = (*Controller)(nil)
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) ListBanners(ctx context.Context, req *apimisc.ListBannersRequest) (*apimisc.ListBannersReply, error) {
	list, err := miscdao.NewBannerDao(nil).List(true)
	if err != nil {
		logger.Errorf("ListBanners err=%v", err)
		return nil, err
	}
	return &apimisc.ListBannersReply{List: toBannerInfos(list)}, nil
}

func (c *Controller) ListAnnouncements(ctx context.Context, req *apimisc.ListAnnouncementsRequest) (*apimisc.ListAnnouncementsReply, error) {
	list, err := miscdao.NewAnnouncementDao(nil).List(true)
	if err != nil {
		logger.Errorf("ListAnnouncements err=%v", err)
		return nil, err
	}
	return &apimisc.ListAnnouncementsReply{List: toAnnouncementInfos(list)}, nil
}

func (c *Controller) AdminCreateBanner(ctx context.Context, req *apimisc.AdminCreateBannerRequest) (*apimisc.AdminBannerReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminCreateBanner RequireAdmin err=%v", err)
		return nil, err
	}
	b, err := miscdao.NewBannerDao(nil).Create(mediastore.Get().Normalize(req.ImageUrl), req.Sort, int32(req.Status))
	if err != nil {
		logger.Errorf("AdminCreateBanner err=%v", err)
		return nil, err
	}
	return &apimisc.AdminBannerReply{Banner: toBannerInfo(b)}, nil
}

func (c *Controller) AdminUpdateBanner(ctx context.Context, req *apimisc.AdminUpdateBannerRequest) (*apimisc.AdminBannerReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminUpdateBanner RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 {
		logger.Errorf("AdminUpdateBanner invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	b, err := miscdao.NewBannerDao(nil).Update(req.Id, mediastore.Get().Normalize(req.ImageUrl), req.Sort, int32(req.Status))
	if err != nil {
		logger.Errorf("AdminUpdateBanner id=%d err=%v", req.Id, err)
		return nil, err
	}
	return &apimisc.AdminBannerReply{Banner: toBannerInfo(b)}, nil
}

func (c *Controller) AdminListBanners(ctx context.Context, req *apimisc.AdminListBannersRequest) (*apimisc.ListBannersReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminListBanners RequireAdmin err=%v", err)
		return nil, err
	}
	list, err := miscdao.NewBannerDao(nil).List(false)
	if err != nil {
		logger.Errorf("AdminListBanners err=%v", err)
		return nil, err
	}
	return &apimisc.ListBannersReply{List: toBannerInfos(list)}, nil
}

func (c *Controller) AdminCreateAnnouncement(ctx context.Context, req *apimisc.AdminCreateAnnouncementRequest) (*apimisc.AdminAnnouncementReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminCreateAnnouncement RequireAdmin err=%v", err)
		return nil, err
	}
	a, err := miscdao.NewAnnouncementDao(nil).Create(req.Content, req.Sort, int32(req.Status))
	if err != nil {
		logger.Errorf("AdminCreateAnnouncement err=%v", err)
		return nil, err
	}
	return &apimisc.AdminAnnouncementReply{Announcement: toAnnouncementInfo(a)}, nil
}

func (c *Controller) AdminUpdateAnnouncement(ctx context.Context, req *apimisc.AdminUpdateAnnouncementRequest) (*apimisc.AdminAnnouncementReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminUpdateAnnouncement RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 {
		logger.Errorf("AdminUpdateAnnouncement invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	a, err := miscdao.NewAnnouncementDao(nil).Update(req.Id, req.Content, req.Sort, int32(req.Status))
	if err != nil {
		logger.Errorf("AdminUpdateAnnouncement id=%d err=%v", req.Id, err)
		return nil, err
	}
	return &apimisc.AdminAnnouncementReply{Announcement: toAnnouncementInfo(a)}, nil
}

func (c *Controller) AdminListAnnouncements(ctx context.Context, req *apimisc.AdminListAnnouncementsRequest) (*apimisc.ListAnnouncementsReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminListAnnouncements RequireAdmin err=%v", err)
		return nil, err
	}
	list, err := miscdao.NewAnnouncementDao(nil).List(false)
	if err != nil {
		logger.Errorf("AdminListAnnouncements err=%v", err)
		return nil, err
	}
	return &apimisc.ListAnnouncementsReply{List: toAnnouncementInfos(list)}, nil
}

func toBannerInfo(b *miscmodel.Banner) *apimisc.BannerInfo {
	if b == nil {
		return nil
	}
	return &apimisc.BannerInfo{
		Id:       b.ID,
		ImageUrl: mediastore.Get().Resolve(b.ImageURL),
		Sort:     b.Sort,
		Status:   apimisc.MiscStatus(b.Status),
		Ct:       b.Ct,
		Ut:       b.Ut,
	}
}

func toBannerInfos(list []miscmodel.Banner) []*apimisc.BannerInfo {
	out := make([]*apimisc.BannerInfo, 0, len(list))
	for i := range list {
		out = append(out, toBannerInfo(&list[i]))
	}
	return out
}

func toAnnouncementInfo(a *miscmodel.Announcement) *apimisc.AnnouncementInfo {
	if a == nil {
		return nil
	}
	return &apimisc.AnnouncementInfo{
		Id:      a.ID,
		Content: a.Content,
		Sort:    a.Sort,
		Status:  apimisc.MiscStatus(a.Status),
		Ct:      a.Ct,
		Ut:      a.Ut,
	}
}

func toAnnouncementInfos(list []miscmodel.Announcement) []*apimisc.AnnouncementInfo {
	out := make([]*apimisc.AnnouncementInfo, 0, len(list))
	for i := range list {
		out = append(out, toAnnouncementInfo(&list[i]))
	}
	return out
}
