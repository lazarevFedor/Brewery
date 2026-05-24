const BASE = import.meta.env.VITE_API_URL ?? ''

async function request(method, path, body) {
  const res = await fetch(BASE + path, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  })

  if (!res.ok) {
    let err
    try {
      err = await res.json()
    } catch {
      err = {}
    }
    throw new Error(err.message || err.error || `HTTP ${res.status}`)
  }

  if (res.status === 204) return null

  try {
    return await res.json()
  } catch {
    return null
  }
}

export const http = {
  get: (path) => request('GET', path),
  post: (path, body) => request('POST', path, body),
  put: (path, body) => request('PUT', path, body),
  patch: (path, body) => request('PATCH', path, body),
  delete: (path) => request('DELETE', path),
}

// Фильтры для каталога
export function buildFilters(params) {
  const filters = []
  if (params.abv_min !== undefined) filters.push(`abv:gte:${params.abv_min}`)
  if (params.abv_max !== undefined) filters.push(`abv:lte:${params.abv_max}`)
  if (params.ibu_min !== undefined) filters.push(`ibu:gte:${params.ibu_min}`)
  if (params.ibu_max !== undefined) filters.push(`ibu:lte:${params.ibu_max}`)
  if (params.countries?.length) filters.push(`country:in:${params.countries.join(',')}`)
  if (params.styles?.length) filters.push(`style:in:${params.styles.join(',')}`)
  return filters
}

// Авторизация
export const auth = {
  login: (password) => {
    if (password === 'admin123') {
      localStorage.setItem('admin_token', 'dummy-token')
      return Promise.resolve(true)
    }
    return Promise.reject(new Error('Неверный пароль'))
  },
  logout: () => localStorage.removeItem('admin_token'),
  isAuthenticated: () => localStorage.getItem('admin_token') !== null,
}

// Категории
export const categoriesApi = {
  getAll: () => http.get('/categories'),
  getById: (id) => http.get(`/categories/${id}`),
  create: (data) => http.post('/categories', data),
  update: (id, data) => http.patch(`/categories/${id}`, data),
  delete: (id) => http.delete(`/categories/${id}`),
}

// ENUM классы
export const enumApi = {
  getAll: (entityName = '', fieldName = '') =>
      http.get(`/enums?entity_name=${entityName}&field_name=${fieldName}`),
  create: (data) => http.post('/enums', data),
  update: (id, data) => http.patch(`/enums/${id}`, data),
  delete: (id) => http.delete(`/enums/${id}`),
}

// ENUM значения
export const enumValueApi = {
  getAll: (entityName = '', fieldName = '', enumType = '') =>
      http.get(`/enums/value?entity_name=${entityName}&field_name=${fieldName}&enum_type=${enumType}`),
  create: (data) => http.post('/enums/value', data),
  update: (id, data) => http.patch(`/enums/value/${id}`, data),
  delete: (id) => http.delete(`/enums/value/${id}`),
}

// Параметры категорий
export const parametersApi = {
  getAll: (categoryId = null, type = null) => {
    let url = '/categories/parameters'
    const params = new URLSearchParams()
    if (categoryId) params.append('category_id', categoryId)
    if (type) params.append('type', type)
    const query = params.toString()
    if (query) url += `?${query}`
    return http.get(url)
  },
  createNumeric: (data) => http.post('/categories/parameters/numeric', data),
  createEnum: (data) => http.post('/categories/parameters/enum', data),
  deleteNumeric: (id) => http.delete(`/categories/parameters/${id}?type=numeric`),
  deleteEnum: (id) => http.delete(`/categories/parameters/${id}?type=enum`),
}

// Агрегаты
export const aggregatesApi = {
  getAll: (name = '') => http.get(`/aggregates?name=${name}`),
  getById: (id) => http.get(`/aggregates/${id}`),
  create: (data) => http.post('/aggregates', data),
  update: (id, data) => http.patch(`/aggregates/${id}`, data),
  delete: (id) => http.delete(`/aggregates/${id}`),
  applyToCategory: (categoryId, aggregateId) =>
      http.patch(`/aggregates/${categoryId}/apply?aggregate_id=${aggregateId}`),
}