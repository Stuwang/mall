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

type PostLikeDao struct {
	db *gorm.DB
}

func NewPostLikeDao(db *gorm.DB) *PostLikeDao {
	return &PostLikeDao{db: defaultDB(db)}
}

func (d *PostLikeDao) WithTx(tx *gorm.DB) *PostLikeDao {
	return &PostLikeDao{db: defaultDB(tx)}
}

func (d *PostLikeDao) Get(postID, userID int64) (*model.PostLike, error) {
	var row model.PostLike
	err := d.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("PostLikeDao.Get postID=%d userID=%d err=%v", postID, userID, err)
		return nil, codeerror.OrSystem(err, "PostLikeDao.Get postID=%d userID=%d", postID, userID)
	}
	return &row, nil
}

func (d *PostLikeDao) Create(row *model.PostLike) error {
	if err := d.db.Create(row).Error; err != nil {
		logger.Errorf("PostLikeDao.Create postID=%d userID=%d err=%v", row.PostID, row.UserID, err)
		return codeerror.OrSystem(err, "PostLikeDao.Create postID=%d", row.PostID)
	}
	return nil
}

// Revive 把已删除的点赞行改回未删除。
func (d *PostLikeDao) Revive(postID, userID int64) error {
	now := time.Now().Unix()
	res := d.db.Model(&model.PostLike{}).
		Where("post_id = ? AND user_id = ? AND deleted = ?", postID, userID, kitcommon.DeletedYes).
		Updates(map[string]interface{}{
			"deleted": kitcommon.DeletedNo,
			"ut":      now,
		})
	if res.Error != nil {
		logger.Errorf("PostLikeDao.Revive postID=%d userID=%d err=%v", postID, userID, res.Error)
		return codeerror.OrSystem(res.Error, "PostLikeDao.Revive postID=%d", postID)
	}
	return nil
}

func (d *PostLikeDao) SoftDelete(postID, userID int64) error {
	now := time.Now().Unix()
	res := d.db.Model(&model.PostLike{}).
		Where("post_id = ? AND user_id = ? AND deleted = ?", postID, userID, kitcommon.DeletedNo).
		Updates(map[string]interface{}{
			"deleted": kitcommon.DeletedYes,
			"ut":      now,
		})
	if res.Error != nil {
		logger.Errorf("PostLikeDao.SoftDelete postID=%d userID=%d err=%v", postID, userID, res.Error)
		return codeerror.OrSystem(res.Error, "PostLikeDao.SoftDelete postID=%d", postID)
	}
	return nil
}

// ListLikedPostIDs 当前用户对给定帖子里哪些已赞（未删除）。
func (d *PostLikeDao) ListLikedPostIDs(userID int64, postIDs []int64) (map[int64]bool, error) {
	out := make(map[int64]bool, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}
	var ids []int64
	err := d.db.Model(&model.PostLike{}).
		Where("user_id = ? AND deleted = ? AND post_id IN ?", userID, kitcommon.DeletedNo, postIDs).
		Pluck("post_id", &ids).Error
	if err != nil {
		logger.Errorf("PostLikeDao.ListLikedPostIDs userID=%d err=%v", userID, err)
		return nil, codeerror.OrSystem(err, "PostLikeDao.ListLikedPostIDs userID=%d", userID)
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}
