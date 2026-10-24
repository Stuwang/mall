package miscdao

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	miscmodel "helloworld/pkg/model/misc"
	"mall/kit/codeerror"
)

type AnnouncementDao struct {
	db *gorm.DB
}

func (d *AnnouncementDao) WithTx(tx *gorm.DB) *AnnouncementDao {
	return &AnnouncementDao{db: tx}
}

func (d *AnnouncementDao) Create(content string, sort int32, status int32) (*miscmodel.Announcement, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "AnnouncementDao.Create empty content")
	}
	if status == 0 {
		status = miscmodel.StatusOn
	}
	if status != miscmodel.StatusOn && status != miscmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "AnnouncementDao.Create invalid status=%d", status)
	}
	now := time.Now().Unix()
	a := &miscmodel.Announcement{
		Content: content,
		Sort:    sort,
		Status:  status,
		Ct:      now,
		Ut:      now,
	}
	if err := d.db.Create(a).Error; err != nil {
		return nil, codeerror.OrSystem(err, "AnnouncementDao.Create")
	}
	return a, nil
}

func (d *AnnouncementDao) Update(id int64, content string, sort int32, status int32) (*miscmodel.Announcement, error) {
	a, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "AnnouncementDao.Update GetByID id=%d", id)
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "AnnouncementDao.Update empty content id=%d", id)
	}
	if status != miscmodel.StatusOn && status != miscmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "AnnouncementDao.Update invalid status=%d id=%d", status, id)
	}
	a.Content = content
	a.Sort = sort
	a.Status = status
	a.Ut = time.Now().Unix()
	if err := d.db.Model(a).Select("content", "sort", "status", "ut").Updates(a).Error; err != nil {
		return nil, codeerror.OrSystem(err, "AnnouncementDao.Update id=%d", id)
	}
	return a, nil
}

func (d *AnnouncementDao) GetByID(id int64) (*miscmodel.Announcement, error) {
	var a miscmodel.Announcement
	err := d.db.Where("id = ?", id).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("announcement not found"), "AnnouncementDao.GetByID not found id=%d", id)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "AnnouncementDao.GetByID id=%d", id)
	}
	return &a, nil
}

func (d *AnnouncementDao) List(onlyOn bool) ([]miscmodel.Announcement, error) {
	q := d.db.Model(&miscmodel.Announcement{})
	if onlyOn {
		q = q.Where("status = ?", miscmodel.StatusOn)
	}
	var list []miscmodel.Announcement
	if err := q.Order("sort desc, id asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "AnnouncementDao.List onlyOn=%v", onlyOn)
	}
	return list, nil
}
