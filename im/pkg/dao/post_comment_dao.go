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

type PostCommentDao struct {
	db *gorm.DB
}

func NewPostCommentDao(db *gorm.DB) *PostCommentDao {
	return &PostCommentDao{db: defaultDB(db)}
}

func (d *PostCommentDao) WithTx(tx *gorm.DB) *PostCommentDao {
	return &PostCommentDao{db: defaultDB(tx)}
}

func (d *PostCommentDao) Create(c *model.PostComment) error {
	if err := d.db.Create(c).Error; err != nil {
		logger.Errorf("PostCommentDao.Create postID=%d userID=%d err=%v", c.PostID, c.UserID, err)
		return codeerror.OrSystem(err, "PostCommentDao.Create postID=%d", c.PostID)
	}
	return nil
}

func (d *PostCommentDao) GetByID(id int64) (*model.PostComment, error) {
	var c model.PostComment
	err := d.db.Where("id = ?", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("PostCommentDao.GetByID id=%d err=%v", id, err)
		return nil, codeerror.OrSystem(err, "PostCommentDao.GetByID id=%d", id)
	}
	return &c, nil
}

func (d *PostCommentDao) GetByClientID(postID, userID int64, clientCommentID string) (*model.PostComment, error) {
	var c model.PostComment
	err := d.db.Where(
		"post_id = ? AND user_id = ? AND client_comment_id = ?",
		postID, userID, clientCommentID,
	).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("PostCommentDao.GetByClientID postID=%d userID=%d err=%v", postID, userID, err)
		return nil, codeerror.OrSystem(err, "PostCommentDao.GetByClientID postID=%d", postID)
	}
	return &c, nil
}

// ListActive 按 id 升序。afterID > 0 时只返回更大的 id。
func (d *PostCommentDao) ListActive(postID, afterID int64, limit int) ([]*model.PostComment, error) {
	q := d.db.Where("post_id = ? AND deleted = ?", postID, kitcommon.DeletedNo)
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	var list []*model.PostComment
	err := q.Order("id ASC").Limit(limit).Find(&list).Error
	if err != nil {
		logger.Errorf("PostCommentDao.ListActive postID=%d err=%v", postID, err)
		return nil, codeerror.OrSystem(err, "PostCommentDao.ListActive postID=%d", postID)
	}
	return list, nil
}

func (d *PostCommentDao) SoftDelete(userID, commentID int64) (bool, error) {
	now := time.Now().Unix()
	res := d.db.Model(&model.PostComment{}).
		Where("id = ? AND user_id = ? AND deleted = ?", commentID, userID, kitcommon.DeletedNo).
		Updates(map[string]interface{}{
			"deleted": kitcommon.DeletedYes,
			"ut":      now,
		})
	if res.Error != nil {
		logger.Errorf("PostCommentDao.SoftDelete userID=%d commentID=%d err=%v", userID, commentID, res.Error)
		return false, codeerror.OrSystem(res.Error, "PostCommentDao.SoftDelete commentID=%d", commentID)
	}
	return res.RowsAffected > 0, nil
}
