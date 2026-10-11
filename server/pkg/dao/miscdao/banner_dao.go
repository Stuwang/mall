package miscdao

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	miscmodel "helloworld/pkg/model/misc"
)

type BannerDao struct {
	db *gorm.DB
}

func (d *BannerDao) WithTx(tx *gorm.DB) *BannerDao {
	return &BannerDao{db: tx}
}

func (d *BannerDao) Create(imageURL string, sort int32, status int32) (*miscmodel.Banner, error) {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "BannerDao.Create empty image_url")
	}
	if status == 0 {
		status = miscmodel.StatusOn
	}
	if status != miscmodel.StatusOn && status != miscmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "BannerDao.Create invalid status=%d", status)
	}
	now := time.Now().Unix()
	b := &miscmodel.Banner{
		ImageURL: imageURL,
		Sort:     sort,
		Status:   status,
		Ct:       now,
		Ut:       now,
	}
	if err := d.db.Create(b).Error; err != nil {
		return nil, codeerror.OrSystem(err, "BannerDao.Create")
	}
	return b, nil
}

func (d *BannerDao) Update(id int64, imageURL string, sort int32, status int32) (*miscmodel.Banner, error) {
	b, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "BannerDao.Update GetByID id=%d", id)
	}
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "BannerDao.Update empty image_url id=%d", id)
	}
	if status != miscmodel.StatusOn && status != miscmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "BannerDao.Update invalid status=%d id=%d", status, id)
	}
	b.ImageURL = imageURL
	b.Sort = sort
	b.Status = status
	b.Ut = time.Now().Unix()
	if err := d.db.Model(b).Select("image_url", "sort", "status", "ut").Updates(b).Error; err != nil {
		return nil, codeerror.OrSystem(err, "BannerDao.Update id=%d", id)
	}
	return b, nil
}

func (d *BannerDao) GetByID(id int64) (*miscmodel.Banner, error) {
	var b miscmodel.Banner
	err := d.db.Where("id = ?", id).First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("banner not found"), "BannerDao.GetByID not found id=%d", id)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "BannerDao.GetByID id=%d", id)
	}
	return &b, nil
}

func (d *BannerDao) List(onlyOn bool) ([]miscmodel.Banner, error) {
	q := d.db.Model(&miscmodel.Banner{})
	if onlyOn {
		q = q.Where("status = ?", miscmodel.StatusOn)
	}
	var list []miscmodel.Banner
	if err := q.Order("sort desc, id asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "BannerDao.List onlyOn=%v", onlyOn)
	}
	return list, nil
}
