package productdao

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/pkg/model/common"
	productmodel "helloworld/pkg/model/product"
)

type ImageDao struct {
	db *gorm.DB
}

func (d *ImageDao) WithTx(tx *gorm.DB) *ImageDao {
	return &ImageDao{db: tx}
}

func (d *ImageDao) ListByProductID(productID int64) ([]productmodel.ProductImage, error) {
	var list []productmodel.ProductImage
	err := d.db.Where("product_id = ? AND deleted = ?", productID, common.DeletedNo).
		Order("sort asc, id asc").Find(&list).Error
	if err != nil {
		return nil, codeerror.OrSystem(err, "ImageDao.ListByProductID productID=%d", productID)
	}
	return list, nil
}

func (d *ImageDao) ListByProductIDs(productIDs []int64) (map[int64][]productmodel.ProductImage, error) {
	out := make(map[int64][]productmodel.ProductImage, len(productIDs))
	if len(productIDs) == 0 {
		return out, nil
	}
	var list []productmodel.ProductImage
	err := d.db.Where("product_id IN ? AND deleted = ?", productIDs, common.DeletedNo).
		Order("sort asc, id asc").Find(&list).Error
	if err != nil {
		return nil, codeerror.OrSystem(err, "ImageDao.ListByProductIDs count=%d", len(productIDs))
	}
	for i := range list {
		pid := list[i].ProductID
		out[pid] = append(out[pid], list[i])
	}
	return out, nil
}

// Replace 全量替换商品图片（软删除旧图后新增）
func (d *ImageDao) Replace(productID int64, urls []string) error {
	if productID <= 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "ImageDao.Replace invalid productID=%d", productID)
	}
	now := time.Now().Unix()
	if err := d.db.Model(&productmodel.ProductImage{}).
		Where("product_id = ? AND deleted = ?", productID, common.DeletedNo).
		Updates(map[string]interface{}{
			"deleted": common.DeletedYes,
		}).Error; err != nil {
		return codeerror.OrSystem(err, "ImageDao.Replace soft delete productID=%d", productID)
	}

	var rows []productmodel.ProductImage
	for i, raw := range urls {
		url := strings.TrimSpace(raw)
		if url == "" {
			continue
		}
		rows = append(rows, productmodel.ProductImage{
			ProductID: productID,
			URL:       url,
			Sort:      int32(i),
			Deleted:   common.DeletedNo,
			Ct:        now,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	if err := d.db.Create(&rows).Error; err != nil {
		return codeerror.OrSystem(err, "ImageDao.Replace create productID=%d count=%d", productID, len(rows))
	}
	return nil
}
