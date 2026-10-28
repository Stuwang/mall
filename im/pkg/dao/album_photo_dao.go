package imdao

import (
	"errors"
	"time"

	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"
	kitcommon "mall/kit/model/common"

	"gorm.io/gorm"
)

type AlbumPhotoDao struct {
	db *gorm.DB
}

func NewAlbumPhotoDao(db *gorm.DB) *AlbumPhotoDao {
	return &AlbumPhotoDao{db: defaultDB(db)}
}

func (d *AlbumPhotoDao) WithTx(tx *gorm.DB) *AlbumPhotoDao {
	return &AlbumPhotoDao{db: defaultDB(tx)}
}

func (d *AlbumPhotoDao) Create(p *model.AlbumPhoto) error {
	if err := d.db.Create(p).Error; err != nil {
		logger.Errorf("AlbumPhotoDao.Create userID=%d err=%v", p.UserID, err)
		return codeerror.OrSystem(err, "AlbumPhotoDao.Create userID=%d", p.UserID)
	}
	return nil
}

func (d *AlbumPhotoDao) GetByID(id int64) (*model.AlbumPhoto, error) {
	var p model.AlbumPhoto
	err := d.db.Where("id = ?", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("AlbumPhotoDao.GetByID id=%d err=%v", id, err)
		return nil, codeerror.OrSystem(err, "AlbumPhotoDao.GetByID id=%d", id)
	}
	return &p, nil
}

// CountActive 未删除照片张数。
func (d *AlbumPhotoDao) CountActive(userID int64) (int64, error) {
	var n int64
	err := d.db.Model(&model.AlbumPhoto{}).
		Where("user_id = ? AND deleted = ?", userID, kitcommon.DeletedNo).
		Count(&n).Error
	if err != nil {
		logger.Errorf("AlbumPhotoDao.CountActive userID=%d err=%v", userID, err)
		return 0, codeerror.OrSystem(err, "AlbumPhotoDao.CountActive userID=%d", userID)
	}
	return n, nil
}

// ListActive 按 id 降序。beforeID > 0 时只返回更小的 id。limit 由调用方约束。
func (d *AlbumPhotoDao) ListActive(userID, beforeID int64, limit int) ([]*model.AlbumPhoto, error) {
	q := d.db.Where("user_id = ? AND deleted = ?", userID, kitcommon.DeletedNo)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	var list []*model.AlbumPhoto
	err := q.Order("id DESC").Limit(limit).Find(&list).Error
	if err != nil {
		logger.Errorf("AlbumPhotoDao.ListActive userID=%d beforeID=%d err=%v", userID, beforeID, err)
		return nil, codeerror.OrSystem(err, "AlbumPhotoDao.ListActive userID=%d", userID)
	}
	return list, nil
}

// MapLatestByUserIDs 批量取每个用户最新 limitPerUser 张未删照片（按 id 降序）。
func (d *AlbumPhotoDao) MapLatestByUserIDs(userIDs []int64, limitPerUser int) (map[int64][]*model.AlbumPhoto, error) {
	out := make(map[int64][]*model.AlbumPhoto, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	if limitPerUser <= 0 {
		limitPerUser = 4
	}
	for _, uid := range userIDs {
		out[uid] = nil
	}
	// 在线人数通常很少；按人限量查，避免一次拉全量相册。
	for _, uid := range userIDs {
		list, err := d.ListActive(uid, 0, limitPerUser)
		if err != nil {
			logger.Errorf("AlbumPhotoDao.MapLatestByUserIDs userID=%d err=%v", uid, err)
			return nil, err
		}
		out[uid] = list
	}
	return out, nil
}

// SoftDelete 仅当属于 userID 且未删除时改标记。返回是否改到一行。
func (d *AlbumPhotoDao) SoftDelete(userID, photoID int64) (bool, error) {
	now := time.Now().Unix()
	res := d.db.Model(&model.AlbumPhoto{}).
		Where("id = ? AND user_id = ? AND deleted = ?", photoID, userID, kitcommon.DeletedNo).
		Updates(map[string]interface{}{
			"deleted": kitcommon.DeletedYes,
			"ut":      now,
		})
	if res.Error != nil {
		logger.Errorf("AlbumPhotoDao.SoftDelete userID=%d photoID=%d err=%v", userID, photoID, res.Error)
		return false, codeerror.OrSystem(res.Error, "AlbumPhotoDao.SoftDelete userID=%d photoID=%d", userID, photoID)
	}
	return res.RowsAffected > 0, nil
}
