package imdao

import (
	"errors"
	"time"

	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"
	kitcommon "mall/kit/model/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostDao struct {
	db *gorm.DB
}

func NewPostDao(db *gorm.DB) *PostDao {
	return &PostDao{db: defaultDB(db)}
}

func (d *PostDao) WithTx(tx *gorm.DB) *PostDao {
	return &PostDao{db: defaultDB(tx)}
}

func (d *PostDao) Create(p *model.Post) error {
	if err := d.db.Create(p).Error; err != nil {
		logger.Errorf("PostDao.Create userID=%d err=%v", p.UserID, err)
		return codeerror.OrSystem(err, "PostDao.Create userID=%d", p.UserID)
	}
	return nil
}

func (d *PostDao) GetByID(id int64) (*model.Post, error) {
	var p model.Post
	err := d.db.Where("id = ?", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("PostDao.GetByID id=%d err=%v", id, err)
		return nil, codeerror.OrSystem(err, "PostDao.GetByID id=%d", id)
	}
	return &p, nil
}

func (d *PostDao) GetByIDForUpdate(id int64) (*model.Post, error) {
	var p model.Post
	err := d.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("PostDao.GetByIDForUpdate id=%d err=%v", id, err)
		return nil, codeerror.OrSystem(err, "PostDao.GetByIDForUpdate id=%d", id)
	}
	return &p, nil
}

func (d *PostDao) GetByClientID(userID int64, clientPostID string) (*model.Post, error) {
	var p model.Post
	err := d.db.Where("user_id = ? AND client_post_id = ?", userID, clientPostID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("PostDao.GetByClientID userID=%d key=%s err=%v", userID, clientPostID, err)
		return nil, codeerror.OrSystem(err, "PostDao.GetByClientID userID=%d", userID)
	}
	return &p, nil
}

// ListRecommend 按 recommend_score、id 降序。hasCursor 为真时取更小的游标。
func (d *PostDao) ListRecommend(hasCursor bool, cursorScore, cursorID int64, limit int) ([]*model.Post, error) {
	q := d.db.Where("deleted = ?", kitcommon.DeletedNo)
	if hasCursor {
		q = q.Where(
			"(recommend_score < ?) OR (recommend_score = ? AND id < ?)",
			cursorScore, cursorScore, cursorID,
		)
	}
	var list []*model.Post
	err := q.Order("recommend_score DESC, id DESC").Limit(limit).Find(&list).Error
	if err != nil {
		logger.Errorf("PostDao.ListRecommend err=%v", err)
		return nil, codeerror.OrSystem(err, "PostDao.ListRecommend")
	}
	return list, nil
}

func (d *PostDao) SoftDelete(userID, postID int64) (bool, error) {
	now := time.Now().Unix()
	res := d.db.Model(&model.Post{}).
		Where("id = ? AND user_id = ? AND deleted = ?", postID, userID, kitcommon.DeletedNo).
		Updates(map[string]interface{}{
			"deleted": kitcommon.DeletedYes,
			"ut":      now,
		})
	if res.Error != nil {
		logger.Errorf("PostDao.SoftDelete userID=%d postID=%d err=%v", userID, postID, res.Error)
		return false, codeerror.OrSystem(res.Error, "PostDao.SoftDelete postID=%d", postID)
	}
	return res.RowsAffected > 0, nil
}

// IncLike 点赞数 +1，并重算推荐分。
func (d *PostDao) IncLike(postID int64) error {
	now := time.Now().Unix()
	res := d.db.Model(&model.Post{}).
		Where("id = ? AND deleted = ?", postID, kitcommon.DeletedNo).
		Updates(map[string]interface{}{
			"like_count":      gorm.Expr("like_count + 1"),
			"recommend_score": gorm.Expr("like_count + 1 + comment_count"),
			"ut":              now,
		})
	if res.Error != nil {
		logger.Errorf("PostDao.IncLike postID=%d err=%v", postID, res.Error)
		return codeerror.OrSystem(res.Error, "PostDao.IncLike postID=%d", postID)
	}
	return nil
}

// DecLike 点赞数 -1（仅当 >0），并重算推荐分。
func (d *PostDao) DecLike(postID int64) error {
	now := time.Now().Unix()
	res := d.db.Model(&model.Post{}).
		Where("id = ? AND deleted = ? AND like_count > 0", postID, kitcommon.DeletedNo).
		Updates(map[string]interface{}{
			"like_count":      gorm.Expr("like_count - 1"),
			"recommend_score": gorm.Expr("like_count - 1 + comment_count"),
			"ut":              now,
		})
	if res.Error != nil {
		logger.Errorf("PostDao.DecLike postID=%d err=%v", postID, res.Error)
		return codeerror.OrSystem(res.Error, "PostDao.DecLike postID=%d", postID)
	}
	return nil
}

func (d *PostDao) IncComment(postID int64) error {
	now := time.Now().Unix()
	res := d.db.Model(&model.Post{}).
		Where("id = ? AND deleted = ?", postID, kitcommon.DeletedNo).
		Updates(map[string]interface{}{
			"comment_count":   gorm.Expr("comment_count + 1"),
			"recommend_score": gorm.Expr("like_count + comment_count + 1"),
			"ut":              now,
		})
	if res.Error != nil {
		logger.Errorf("PostDao.IncComment postID=%d err=%v", postID, res.Error)
		return codeerror.OrSystem(res.Error, "PostDao.IncComment postID=%d", postID)
	}
	return nil
}

func (d *PostDao) DecComment(postID int64) error {
	now := time.Now().Unix()
	res := d.db.Model(&model.Post{}).
		Where("id = ? AND deleted = ? AND comment_count > 0", postID, kitcommon.DeletedNo).
		Updates(map[string]interface{}{
			"comment_count":   gorm.Expr("comment_count - 1"),
			"recommend_score": gorm.Expr("like_count + comment_count - 1"),
			"ut":              now,
		})
	if res.Error != nil {
		logger.Errorf("PostDao.DecComment postID=%d err=%v", postID, res.Error)
		return codeerror.OrSystem(res.Error, "PostDao.DecComment postID=%d", postID)
	}
	return nil
}
