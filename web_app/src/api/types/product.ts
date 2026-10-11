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

export interface ListProductsRequest extends WithBaseParam, PageQuery {
  keyword?: string
}

export interface ListProductsByCategoryRequest extends WithBaseParam, PageQuery {
  category_id: number
}

export interface ListProductsReply {
  list: ProductInfo[]
  total: number
}

export interface GetProductRequest extends WithBaseParam {
  id: number
}

export interface GetProductReply {
  product: ProductInfo
}

export type ListCategoriesRequest = WithBaseParam

export interface ListCategoriesReply {
  list: CategoryInfo[]
}
