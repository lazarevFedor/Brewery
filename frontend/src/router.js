// src/router.js
import { renderCatalog } from './pages/catalog.jsx'
import { renderDetail } from './pages/beerDetail.jsx'
// import { renderSearch } from './pages/search.js'
// import { renderAdmin } from './pages/admin.js'

const routes = {
  '/': renderCatalog,
  '/beer/:id': renderDetail,
  // '/search': renderSearch,
  // '/admin': renderAdmin,
}

export function initRouter() {
  handleRoute(window.location.pathname)

  window.addEventListener('popstate', () => {
    handleRoute(window.location.pathname)
  })
}

export function navigate(path) {
  window.history.pushState({}, '', path)
  handleRoute(path)
}

function handleRoute(pathname) {
  for (const [pattern, handler] of Object.entries(routes)) {
    const match = matchPath(pattern, pathname)
    if (match) {
      const container = document.getElementById('app')
      handler(container, match.params)
      window.scrollTo(0, 0)
      return
    }
  }

  // 404
  document.getElementById('app').innerHTML = '<h1>404 — страница не найдена</h1>'
}

function matchPath(pattern, pathname) {
  const patternParts = pattern.split('/').filter(Boolean)
  const pathParts = pathname.split('/').filter(Boolean)

  if (patternParts.length !== pathParts.length) return null

  const params = {}
  for (let i = 0; i < patternParts.length; i++) {
    const pp = patternParts[i]
    const ap = pathParts[i]

    if (pp.startsWith(':')) {
      params[pp.slice(1)] = ap
    } else if (pp !== ap) {
      return null
    }
  }

  return { params }
}