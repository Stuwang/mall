package controller

import (
	"context"

	apiprofile "mall/im/api/profile"
	"mall/im/pkg/biz"
	"mall/kit/auth"
	"mall/kit/logger"
)

// ProfileController 主页与相册。
type ProfileController struct{}

func NewProfileController() *ProfileController {
	return &ProfileController{}
}

func (c *ProfileController) GetProfile(ctx context.Context, req *apiprofile.GetProfileRequest) (*apiprofile.GetProfileReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.GetProfile(a.UserID, req.GetUserId())
	if err != nil {
		logger.Errorf("GetProfile viewer=%d target=%d err=%v", a.UserID, req.GetUserId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *ProfileController) ListAlbum(ctx context.Context, req *apiprofile.ListAlbumRequest) (*apiprofile.ListAlbumReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.ListAlbum(a.UserID, req.GetUserId(), req.GetBeforeId(), req.GetLimit())
	if err != nil {
		logger.Errorf("ListAlbum viewer=%d target=%d err=%v", a.UserID, req.GetUserId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *ProfileController) AddAlbumPhoto(ctx context.Context, req *apiprofile.AddAlbumPhotoRequest) (*apiprofile.AddAlbumPhotoReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.AddAlbumPhoto(a.UserID, req.GetImageUri(), req.GetCaption())
	if err != nil {
		logger.Errorf("AddAlbumPhoto user=%d err=%v", a.UserID, err)
		return nil, err
	}
	return reply, nil
}

func (c *ProfileController) DeleteAlbumPhoto(ctx context.Context, req *apiprofile.DeleteAlbumPhotoRequest) (*apiprofile.DeleteAlbumPhotoReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.DeleteAlbumPhoto(a.UserID, req.GetPhotoId())
	if err != nil {
		logger.Errorf("DeleteAlbumPhoto user=%d photoID=%d err=%v", a.UserID, req.GetPhotoId(), err)
		return nil, err
	}
	return reply, nil
}

var _ apiprofile.ProfileHTTPServer = (*ProfileController)(nil)
