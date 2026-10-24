package productcontroller

import (
	apiproduct "helloworld/api/product"
	productmodel "helloworld/pkg/model/product"
	"mall/file/mediastore"
)

func toProductInfo(p *productmodel.Product, images []productmodel.ProductImage) *apiproduct.ProductInfo {
	if p == nil {
		return nil
	}
	store := mediastore.Get()
	urls := make([]string, 0, len(images))
	for i := range images {
		urls = append(urls, store.Resolve(images[i].URL))
	}
	return &apiproduct.ProductInfo{
		Id:          p.ID,
		CategoryId:  p.CategoryID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Status:      p.Status,
		Images:      urls,
		Ct:          p.Ct,
		Ut:          p.Ut,
		IsHot:       p.IsHot,
	}
}

func toProductInfos(list []productmodel.Product, imageMap map[int64][]productmodel.ProductImage) []*apiproduct.ProductInfo {
	out := make([]*apiproduct.ProductInfo, 0, len(list))
	for i := range list {
		out = append(out, toProductInfo(&list[i], imageMap[list[i].ID]))
	}
	return out
}

func toCategoryInfo(c *productmodel.Category) *apiproduct.CategoryInfo {
	if c == nil {
		return nil
	}
	return &apiproduct.CategoryInfo{
		Id:     c.ID,
		Name:   c.Name,
		Sort:   c.Sort,
		Status: c.Status,
		Ct:     c.Ct,
		Ut:     c.Ut,
	}
}

func toCategoryInfos(list []productmodel.Category) []*apiproduct.CategoryInfo {
	out := make([]*apiproduct.CategoryInfo, 0, len(list))
	for i := range list {
		out = append(out, toCategoryInfo(&list[i]))
	}
	return out
}

func productIDs(list []productmodel.Product) []int64 {
	ids := make([]int64, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].ID)
	}
	return ids
}

func normalizeImageURLs(urls []string) []string {
	store := mediastore.Get()
	out := make([]string, 0, len(urls))
	for _, u := range urls {
		n := store.Normalize(u)
		if n == "" {
			continue
		}
		out = append(out, n)
	}
	return out
}
