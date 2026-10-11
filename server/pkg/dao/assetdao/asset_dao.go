package assetdao

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	assetmodel "helloworld/pkg/model/asset"
)

type AssetDao struct {
	db *gorm.DB
}

func (d *AssetDao) WithTx(tx *gorm.DB) *AssetDao {
	return &AssetDao{db: tx}
}

func (d *AssetDao) ListByUserID(userID int64, assetType int32) ([]assetmodel.UserAsset, error) {
	q := d.db.Where("user_id = ?", userID)
	if assetType > 0 {
		q = q.Where("type = ?", assetType)
	}
	var list []assetmodel.UserAsset
	if err := q.Order("type asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "AssetDao.ListByUserID userID=%d type=%d", userID, assetType)
	}
	return list, nil
}

func (d *AssetDao) GetByUserAndType(userID int64, assetType int32) (*assetmodel.UserAsset, error) {
	a, err := d.FindByUserAndType(userID, assetType)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("asset not found"), "AssetDao.GetByUserAndType not found userID=%d type=%d", userID, assetType)
	}
	return a, nil
}

// FindByUserAndType 查询资产；不存在时返回 nil, nil
func (d *AssetDao) FindByUserAndType(userID int64, assetType int32) (*assetmodel.UserAsset, error) {
	var a assetmodel.UserAsset
	err := d.db.Where("user_id = ? AND type = ?", userID, assetType).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "AssetDao.FindByUserAndType userID=%d type=%d", userID, assetType)
	}
	return &a, nil
}

func (d *AssetDao) Create(a *assetmodel.UserAsset) error {
	if a == nil {
		return codeerror.OrSystem(codeerror.ErrorParam, "AssetDao.Create nil")
	}
	if err := d.db.Create(a).Error; err != nil {
		return codeerror.OrSystem(err, "AssetDao.Create userID=%d type=%d", a.UserID, a.Type)
	}
	return nil
}

func (d *AssetDao) UpdateBalance(a *assetmodel.UserAsset) error {
	if a == nil {
		return codeerror.OrSystem(codeerror.ErrorParam, "AssetDao.UpdateBalance nil")
	}
	a.Ut = time.Now().Unix()
	if err := d.db.Model(a).Select("balance", "ut").Updates(a).Error; err != nil {
		return codeerror.OrSystem(err, "AssetDao.UpdateBalance id=%d", a.ID)
	}
	return nil
}
