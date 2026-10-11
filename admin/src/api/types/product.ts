import type { PageQuery, WithBaseParam } from './common'

/** product.v1.ProductStatus */
export const ProductStatus = {
  UNSPECIFIED: 0,
  ON: 1,
  OFF: 2,
} as const
export type ProductStatus = (typeof ProductStatus)[keyof typeof ProductStatus]

/** product.v1.CategoryStatus */
export const CategoryStatus = {
  UNSPECIFIED: 0,
  ON: 1,
  OFF: 2,
} as const
export type CategoryStatus = (typeof CategoryStatus)[keyof typeof CategoryStatus]

export interface CategoryInfo {
  id: number
  name: string
  sort: number
  status: CategoryStatus
  ct: number
  ut: number
}

export interface ProductInfo {
  id: number
  category_id: number
  name: string
  description: string
  /** 价格，单位分 */
  price: number
  stock: number
  status: ProductStatus
  images: string[]
  ct: number
  ut: number
}

export interface ListCategoriesReply {
  list: CategoryInfo[]
}

export interface ListProductsReply {
  list: ProductInfo[]
  total: number
}

export interface GetProductReply {
  product: ProductInfo
}

export interface AdminCreateCategoryRequest extends WithBaseParam {
  name: string
  sort: number
  status: CategoryStatus
}

export interface AdminUpdateCategoryRequest extends WithBaseParam {
  id: number
  name: string
  sort: number
  status: CategoryStatus
}

export type AdminListCategoriesRequest = WithBaseParam

export interface AdminCategoryReply {
  category: CategoryInfo
}

export interface AdminCreateProductRequest extends WithBaseParam {
  category_id: number
  name: string
  description: string
  price: number
  stock: number
  images: string[]
  status: ProductStatus
}

export interface AdminUpdateProductRequest extends WithBaseParam {
  id: number
  name: string
  description: string
  price: number
  images: string[]
}

export interface AdminSetProductStatusRequest extends WithBaseParam {
  id: number
  status: ProductStatus
}

export interface AdminUpdateProductStockRequest extends WithBaseParam {
  id: number
  stock: number
}

export interface AdminUpdateProductCategoryRequest extends WithBaseParam {
  id: number
  category_id: number
}

export interface AdminListProductsRequest extends WithBaseParam, PageQuery {
  category_id?: number
  status?: ProductStatus
  keyword?: string
}

export interface AdminGetProductRequest extends WithBaseParam {
  id: number
}

export interface AdminProductReply {
  product: ProductInfo
}
