package productcontroller

import (
	apiproduct "helloworld/api/product"
	productmodel "helloworld/pkg/model/product"
)

func toProductInfo(p *productmodel.Product) *apiproduct.ProductInfo {
	if p == nil {
		return nil
	}
	return &apiproduct.ProductInfo{
		Id:          p.ID,
		CategoryId:  p.CategoryID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Status:      apiproduct.ProductStatus(p.Status),
		Cover:       p.Cover,
		Ct:          p.Ct,
		Ut:          p.Ut,
	}
}

func toProductInfos(list []productmodel.Product) []*apiproduct.ProductInfo {
	out := make([]*apiproduct.ProductInfo, 0, len(list))
	for i := range list {
		out = append(out, toProductInfo(&list[i]))
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
		Status: apiproduct.CategoryStatus(c.Status),
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
