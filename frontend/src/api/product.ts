import request from '../utils/request'

export interface ProductLine {
  id: number
  name: string
  description?: string
  status: number
  created_at?: string
  updated_at?: string
  products?: Product[]
}

export interface Product {
  id: number
  name: string
  code: string
  description?: string
  status: number
  product_line_id: number
  product_line?: ProductLine
  created_at?: string
  updated_at?: string
}

export interface ProductListResponse {
  list: Product[]
  total: number
  page: number
  size: number
}

export interface CreateProductLineRequest {
  name: string
  description?: string
  status?: number
}

export interface CreateProductRequest {
  name: string
  code: string
  description?: string
  status?: number
  product_line_id: number
}

// 产品线相关API
export const getProductLines = async (params?: {
  keyword?: string
  status?: number
}): Promise<ProductLine[]> => {
  return request.get('/product-lines', { params })
}

export const getProductLine = async (id: number): Promise<ProductLine> => {
  return request.get(`/product-lines/${id}`)
}

export const createProductLine = async (data: CreateProductLineRequest): Promise<ProductLine> => {
  return request.post('/product-lines', data)
}

export const updateProductLine = async (id: number, data: Partial<CreateProductLineRequest>): Promise<ProductLine> => {
  return request.put(`/product-lines/${id}`, data)
}

export const deleteProductLine = async (id: number): Promise<void> => {
  return request.delete(`/product-lines/${id}`)
}

// 产品相关API
export const getProducts = async (params?: {
  keyword?: string
  product_line_id?: number
  status?: number
  page?: number
  size?: number
}): Promise<ProductListResponse> => {
  return request.get('/products', { params })
}

export const getProduct = async (id: number): Promise<Product> => {
  return request.get(`/products/${id}`)
}

export const createProduct = async (data: CreateProductRequest): Promise<Product> => {
  return request.post('/products', data)
}

export const updateProduct = async (id: number, data: Partial<CreateProductRequest>): Promise<Product> => {
  return request.put(`/products/${id}`, data)
}

export const deleteProduct = async (id: number): Promise<void> => {
  return request.delete(`/products/${id}`)
}

