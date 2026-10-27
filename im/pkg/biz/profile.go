package biz

import (
	"strings"
	"time"
	"unicode/utf8"

	apiprofile "mall/im/api/profile"
	"mall/im/pkg/bizerror"
	imdao "mall/im/pkg/dao"
	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"
	kitcommon "mall/kit/model/common"
	userdao "mall/user/dao"
)

const (
	albumMaxPhotos   = 100
	albumCaptionMax  = 80
	albumListDefault = 30
	albumListMax     = 50
)

// GetProfile 主页。targetUserID <= 0 表示查看自己。
func GetProfile(viewerID, targetUserID int64) (*apiprofile.GetProfileReply, error) {
	uid, err := resolveProfileUser(viewerID, targetUserID)
	if err != nil {
		return nil, err
	}
	u, err := userdao.NewUserDao(nil).GetByID(uid)
	if err != nil {
		logger.Errorf("GetProfile GetByID viewer=%d target=%d err=%v", viewerID, uid, err)
		return nil, err
	}
	if uid != viewerID && u.Banned() {
		logger.Errorf("GetProfile target banned viewer=%d target=%d", viewerID, uid)
		return nil, codeerror.ErrorUserBanned
	}
	n, err := imdao.NewAlbumPhotoDao(nil).CountActive(uid)
	if err != nil {
		logger.Errorf("GetProfile CountActive viewer=%d target=%d err=%v", viewerID, uid, err)
		return nil, err
	}
	return &apiprofile.GetProfileReply{
		UserId:     u.ID,
		Nickname:   u.Nickname,
		Avatar:     avatarURL(u.Avatar),
		PhotoCount: n,
	}, nil
}

// ListAlbum 相册列表。targetUserID <= 0 表示自己。
func ListAlbum(viewerID, targetUserID, beforeID int64, limit int32) (*apiprofile.ListAlbumReply, error) {
	uid, err := resolveProfileUser(viewerID, targetUserID)
	if err != nil {
		return nil, err
	}
	u, err := userdao.NewUserDao(nil).GetByID(uid)
	if err != nil {
		logger.Errorf("ListAlbum GetByID viewer=%d target=%d err=%v", viewerID, uid, err)
		return nil, err
	}
	if uid != viewerID && u.Banned() {
		logger.Errorf("ListAlbum target banned viewer=%d target=%d", viewerID, uid)
		return nil, codeerror.ErrorUserBanned
	}
	lim := int(limit)
	if lim <= 0 {
		lim = albumListDefault
	}
	if lim > albumListMax {
		lim = albumListMax
	}
	rows, err := imdao.NewAlbumPhotoDao(nil).ListActive(uid, beforeID, lim+1)
	if err != nil {
		logger.Errorf("ListAlbum ListActive viewer=%d target=%d err=%v", viewerID, uid, err)
		return nil, err
	}
	hasMore := false
	if len(rows) > lim {
		hasMore = true
		rows = rows[:lim]
	}
	list := make([]*apiprofile.AlbumPhoto, 0, len(rows))
	for _, p := range rows {
		list = append(list, toAlbumPhotoPB(p))
	}
	return &apiprofile.ListAlbumReply{List: list, HasMore: hasMore}, nil
}

// AddAlbumPhoto 添加自己的照片。
func AddAlbumPhoto(userID int64, imageURI, caption string) (*apiprofile.AddAlbumPhotoReply, error) {
	u, err := userdao.NewUserDao(nil).GetByID(userID)
	if err != nil {
		logger.Errorf("AddAlbumPhoto GetByID user=%d err=%v", userID, err)
		return nil, err
	}
	if u.Banned() {
		logger.Errorf("AddAlbumPhoto user banned user=%d", userID)
		return nil, codeerror.ErrorUserBanned
	}
	caption = strings.TrimSpace(caption)
	if utf8.RuneCountInString(caption) > albumCaptionMax {
		logger.Errorf("AddAlbumPhoto caption too long user=%d n=%d", userID, utf8.RuneCountInString(caption))
		return nil, codeerror.ErrorParam
	}
	uri, err := normalizeImageURI(imageURI)
	if err != nil {
		logger.Errorf("AddAlbumPhoto image_uri invalid user=%d err=%v", userID, err)
		return nil, codeerror.ErrorParam
	}
	dao := imdao.NewAlbumPhotoDao(nil)
	n, err := dao.CountActive(userID)
	if err != nil {
		logger.Errorf("AddAlbumPhoto CountActive user=%d err=%v", userID, err)
		return nil, err
	}
	if n >= albumMaxPhotos {
		logger.Errorf("AddAlbumPhoto full user=%d count=%d", userID, n)
		return nil, bizerror.ErrorAlbumFull
	}
	now := time.Now().Unix()
	p := &model.AlbumPhoto{
		UserID:  userID,
		URI:     uri,
		Caption: caption,
		Deleted: kitcommon.DeletedNo,
		Ct:      now,
		Ut:      now,
	}
	if err := dao.Create(p); err != nil {
		logger.Errorf("AddAlbumPhoto Create user=%d err=%v", userID, err)
		return nil, err
	}
	return &apiprofile.AddAlbumPhotoReply{Photo: toAlbumPhotoPB(p)}, nil
}

// DeleteAlbumPhoto 软删除自己的照片。
func DeleteAlbumPhoto(userID, photoID int64) (*apiprofile.DeleteAlbumPhotoReply, error) {
	if photoID <= 0 {
		logger.Errorf("DeleteAlbumPhoto photo_id invalid user=%d photoID=%d", userID, photoID)
		return nil, codeerror.ErrorParam
	}
	u, err := userdao.NewUserDao(nil).GetByID(userID)
	if err != nil {
		logger.Errorf("DeleteAlbumPhoto GetByID user=%d err=%v", userID, err)
		return nil, err
	}
	if u.Banned() {
		logger.Errorf("DeleteAlbumPhoto user banned user=%d", userID)
		return nil, codeerror.ErrorUserBanned
	}
	ok, err := imdao.NewAlbumPhotoDao(nil).SoftDelete(userID, photoID)
	if err != nil {
		logger.Errorf("DeleteAlbumPhoto SoftDelete user=%d photoID=%d err=%v", userID, photoID, err)
		return nil, err
	}
	if !ok {
		logger.Errorf("DeleteAlbumPhoto not found user=%d photoID=%d", userID, photoID)
		return nil, bizerror.ErrorAlbumPhoto
	}
	return &apiprofile.DeleteAlbumPhotoReply{}, nil
}

func resolveProfileUser(viewerID, targetUserID int64) (int64, error) {
	if targetUserID <= 0 {
		return viewerID, nil
	}
	return targetUserID, nil
}

func toAlbumPhotoPB(p *model.AlbumPhoto) *apiprofile.AlbumPhoto {
	if p == nil {
		return nil
	}
	return &apiprofile.AlbumPhoto{
		Id:       p.ID,
		ImageUrl: avatarURL(p.URI),
		Caption:  p.Caption,
		Ct:       p.Ct,
	}
}
