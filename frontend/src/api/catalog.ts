import { request } from './client'
import type { CatalogPage, CatalogSort, Genre } from './types'

export interface CatalogQuery {
  sort?: CatalogSort
  genre?: Genre | ''
  limit?: number
  offset?: number
}

export function getCatalog(query: CatalogQuery = {}): Promise<CatalogPage> {
  const params = new URLSearchParams()

  if (query.sort) params.set('sort', query.sort)
  if (query.genre) params.set('genre', query.genre)
  if (query.limit) params.set('limit', String(query.limit))
  if (query.offset) params.set('offset', String(query.offset))

  const suffix = params.size ? `?${params.toString()}` : ''

  return request<CatalogPage>(`/catalog${suffix}`)
}
