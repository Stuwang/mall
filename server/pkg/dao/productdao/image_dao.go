package productdao

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
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
	if err := d.db.Where("product_id = ?", productID).Order("sort asc, id asc").Find(&list).Error; err != nil {
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
	if err := d.db.Where("product_id IN ?", productIDs).Order("sort asc, id asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ImageDao.ListByProductIDs count=%d", len(productIDs))
	}
	for i := range list {
		pid := list[i].ProductID
		out[pid] = append(out[pid], list[i])
	}
	return out, nil
}

// Replace 全量替换商品图片（先删后插）
func (d *ImageDao) Replace(productID int64, urls []string) error {
	if productID <= 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "ImageDao.Replace invalid productID=%d", productID)
	}
	if err := d.db.Where("product_id = ?", productID).Delete(&productmodel.ProductImage{}).Error; err != nil {
		return codeerror.OrSystem(err, "ImageDao.Replace delete productID=%d", productID)
	}
	now := time.Now().Unix()
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
