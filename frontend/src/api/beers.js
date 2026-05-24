// src/api/beers.js
import { http } from './client.js'

export function getBeers({ categoryId, offset = 0, limit = 50, filters = [] } = {}) {
  const p = new URLSearchParams({ offset, limit })
  if (categoryId) p.set('category_id', categoryId)
  filters.forEach(f => p.append('filter', f))
  return http.get(`/api/beers?${p}`)
}

export const getBeer = (id) => http.get(`/api/beers/${id}`)
export const createBeer = (b) => http.post('/api/beers', b)
export const updateBeer = (id, b) => http.put(`/api/beers/${id}`, b)
export const deleteBeer = (id) => http.delete(`/api/beers/${id}`)

export function getCategories() {
  return http.get('/api/categories')
}

export const getCategoryChildren = (id) => http.get(`/api/categories/children/${id}`)
export const getCategoryParams = (id) => http.get(`/api/category-params?category_id=${id}`)

// Примеры запросов с фильтрами в API-стиле:
// getBeers({ filters: ['abv:gte:5', 'abv:lte:10', 'country:in:Бельгия,Германия'] })