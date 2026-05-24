// src/api/client.js
// Базовый HTTP клиент — оборачивает fetch, переиспользуется везде

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
  delete: (path) => request('DELETE', path),
}

// Утилита для построения фильтров API-стиля
export function buildFilters(params) {
  const filters = []
  
  if (params.abv_min !== undefined) filters.push(`abv:gte:${params.abv_min}`)
  if (params.abv_max !== undefined) filters.push(`abv:lte:${params.abv_max}`)
  if (params.ibu_min !== undefined) filters.push(`ibu:gte:${params.ibu_min}`)
  if (params.ibu_max !== undefined) filters.push(`ibu:lte:${params.ibu_max}`)
  if (params.countries && params.countries.length) {
    filters.push(`country:in:${params.countries.join(',')}`)
  }
  if (params.styles && params.styles.length) {
    filters.push(`style:in:${params.styles.join(',')}`)
  }
  
  return filters
}