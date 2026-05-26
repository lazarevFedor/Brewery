const BASE = import.meta.env.VITE_API_URL ?? '/api'

async function request(method, path, body, token = null) {
  const url = BASE + path
  console.log('Request:', method, url)

  const headers = { 'Content-Type': 'application/json' }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(url, {
    method,
    headers,
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

// Сохранение токена
let adminToken = localStorage.getItem('admin_token')
export const setToken = (token) => {
  adminToken = token
  if (token) localStorage.setItem('admin_token', token)
  else localStorage.removeItem('admin_token')
}
export const getToken = () => adminToken

// Auth
export const auth = {
  login: async (username, password) => {
    const res = await fetch(`${BASE}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error(err.message || 'Ошибка авторизации')
    }
    const data = await res.json()
    if (data.token) {
      setToken(data.token)
      return true
    }
    throw new Error('Токен не получен')
  },
  logout: () => setToken(null),
  isAuthenticated: () => !!getToken(),
}

// Categories
export const categoriesApi = {
  getAll: () => request('GET', '/categories'),
  getById: (id) => request('GET', `/categories/${id}`),
  getChildren: (id) => request('GET', `/categories/children/${id}`),
  getParent: (id) => request('GET', `/categories/parent/${id}`),
  create: (data) => request('POST', '/categories', data, getToken()),
  update: (id, data) => request('PATCH', `/categories/${id}`, data, getToken()),
  delete: (id) => request('DELETE', `/categories/${id}`, null, getToken()),
}

// Beers
export const beersApi = {
  getAll: (offset = 0, limit = 100) => request('GET', `/beers?offset=${offset}&limit=${limit}`),
  getById: (id) => request('GET', `/beers/${id}`),
  getByCategory: (categoryId, offset = 0, limit = 100) =>
      request('GET', `/categories/beers/${categoryId}?offset=${offset}&limit=${limit}`),
  create: (data) => request('POST', '/beers', data, getToken()),
  update: (id, data) => request('PATCH', `/beers/${id}`, data, getToken()),
  delete: (id) => request('DELETE', `/beers/${id}`, null, getToken()),
}

// Reviews
export const reviewsApi = {
  getByBeerId: (beerId) => request('GET', `/reviews/${beerId}`),
  create: (beerId, data) => request('POST', `/reviews/${beerId}`, data),
  update: (id, data) => request('PATCH', `/reviews/${id}`, data, getToken()),
  delete: (id) => request('DELETE', `/reviews/${id}`, null, getToken()),
}

// Остальные API...
export const enumApi = {
  getAll: (entityName = '', fieldName = '') =>
      request('GET', `/enums?entity_name=${entityName}&field_name=${fieldName}`, null, getToken()),
  create: (data) => request('POST', '/enums', data, getToken()),
  update: (id, data) => request('PATCH', `/enums/${id}`, data, getToken()),
  delete: (id) => request('DELETE', `/enums/${id}`, null, getToken()),
}

export const enumValueApi = {
  getAll: (entityName = '', fieldName = '', enumType = '') =>
      request('GET', `/enums/value?entity_name=${entityName}&field_name=${fieldName}&enum_type=${enumType}`, null, getToken()),
  create: (data) => request('POST', '/enums/value', data, getToken()),
  update: (id, data) => request('PATCH', `/enums/value/${id}`, data, getToken()),
  delete: (id) => request('DELETE', `/enums/value/${id}`, null, getToken()),
}

export const parametersApi = {
  getAll: (categoryId = null, type = null) => {
    let url = '/categories/parameters'
    const params = new URLSearchParams()
    if (categoryId) params.append('category_id', categoryId)
    if (type) params.append('type', type)
    const query = params.toString()
    if (query) url += `?${query}`
    return request('GET', url, null, getToken())
  },
  createNumeric: (data) => request('POST', '/categories/parameters/numeric', data, getToken()),
  createEnum: (data) => request('POST', '/categories/parameters/enum', data, getToken()),
  deleteNumeric: (id) => request('DELETE', `/categories/parameters/${id}?type=numeric`, null, getToken()),
  deleteEnum: (id) => request('DELETE', `/categories/parameters/${id}?type=enum`, null, getToken()),
  applyToCategory: (categoryId, numericIds, enumIds) =>
      request('PATCH', `/categories/parameters/apply/${categoryId}`,
          { numeric_param_ids: numericIds, enum_param_ids: enumIds }, getToken()),
}

export const aggregatesApi = {
  getAll: (name = '') => request('GET', `/aggregates?name=${name}`, null, getToken()),
  getById: (id) => request('GET', `/aggregates/${id}`, null, getToken()),
  create: (data) => request('POST', '/aggregates', data, getToken()),
  update: (id, data) => request('PATCH', `/aggregates/${id}`, data, getToken()),
  delete: (id) => request('DELETE', `/aggregates/${id}`, null, getToken()),
  applyToCategory: (categoryId, aggregateId) =>
      request('PATCH', `/aggregates/apply/${categoryId}?aggregate_id=${aggregateId}`, null, getToken()),
}