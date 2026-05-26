import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { beersApi, categoriesApi } from '../api/client'
import "../styles/catalog.css"

export function Catalog() {
  const [beers, setBeers] = useState([])
  const [filteredBeers, setFilteredBeers] = useState([])
  const [loading, setLoading] = useState(true)
  const [categories, setCategories] = useState([])
  const [currentCategory, setCurrentCategory] = useState(null)
  const [categoryChildren, setCategoryChildren] = useState([])
  const [categoryPath, setCategoryPath] = useState([])
  const [countries, setCountries] = useState([])
  const [selectedCountries, setSelectedCountries] = useState([])
  const [styles, setStyles] = useState([])
  const [selectedStyles, setSelectedStyles] = useState([])
  const [filters, setFilters] = useState({ abvMin: 0, abvMax: 20, ibuMin: 0, ibuMax: 120 })

  useEffect(() => { loadAllData() }, [])

  const loadAllData = async () => {
    setLoading(true)
    try {
      const [cats, beersData] = await Promise.all([
        categoriesApi.getAll(),
        beersApi.getAll()
      ])
      setCategories(cats)
      setBeers(beersData)
      setFilteredBeers(beersData)

      // Находим корневую категорию и показываем её детей
      const root = cats.find(c => !c.parent_id || c.parent_id === 0)
      if (root) {
        const childrenOfRoot = cats.filter(c => c.parent_id === root.id)
        setCategoryChildren(childrenOfRoot)
      } else {
        setCategoryChildren([])
      }

      // Уникальные страны и стили
      setCountries([...new Set(beersData.map(b => b.country).filter(Boolean))])
      setSelectedCountries([...new Set(beersData.map(b => b.country).filter(Boolean))])
      setStyles([...new Set(beersData.map(b => b.category?.name).filter(Boolean))])
      setSelectedStyles([...new Set(beersData.map(b => b.category?.name).filter(Boolean))])
    } catch (error) {
      console.error('Ошибка загрузки:', error)
    } finally {
      setLoading(false)
    }
  }

  const selectCategory = async (category) => {
    setLoading(true)
    setCurrentCategory(category)

    try {
      const children = await categoriesApi.getChildren(category.id)
      setCategoryChildren(children)

      // Получаем пиво через API
      const beersByCat = await beersApi.getByCategory(category.id)
      console.log(`Пиво в категории "${category.name}":`, beersByCat)
      setFilteredBeers(beersByCat)

      // Хлебные крошки
      const path = []
      let current = category
      while (current) {
        path.unshift({ id: current.id, name: current.name })
        const parent = categories.find(c => c.id === current.parent_id)
        current = parent
      }
      setCategoryPath(path)
    } catch (error) {
      console.error('Ошибка выбора категории:', error)
    } finally {
      setLoading(false)
    }
  }

  const goToRoot = () => {
    const root = categories.find(c => !c.parent_id || c.parent_id === 0)
    if (root) {
      const childrenOfRoot = categories.filter(c => c.parent_id === root.id)
      setCategoryChildren(childrenOfRoot)
    }
    setCurrentCategory(null)
    setCategoryPath([])
    setFilteredBeers(beers)
  }

  const goBack = async () => {
    if (categoryPath.length === 0) {
      goToRoot()
    } else if (categoryPath.length === 1) {
      goToRoot()
    } else {
      const parent = categoryPath[categoryPath.length - 2]
      await selectCategory(parent)
    }
  }

  const applyFilters = () => {
    let filtered = [...beers]
    if (currentCategory) {
      filtered = filtered.filter(b => b.category?.id === currentCategory.id || b.category_id === currentCategory.id)
    }
    filtered = filtered.filter(b => (b.abv || 0) >= filters.abvMin && (b.abv || 0) <= filters.abvMax)
    filtered = filtered.filter(b => (b.ibu || 0) >= filters.ibuMin && (b.ibu || 0) <= filters.ibuMax)
    if (selectedCountries.length) filtered = filtered.filter(b => selectedCountries.includes(b.country))
    if (selectedStyles.length) filtered = filtered.filter(b => selectedStyles.includes(b.category?.name))
    setFilteredBeers(filtered)
  }

  useEffect(() => { applyFilters() }, [beers, currentCategory, selectedCountries, selectedStyles, filters])

  const resetFilters = () => {
    setFilters({ abvMin: 0, abvMax: 20, ibuMin: 0, ibuMax: 120 })
    setSelectedCountries([...countries])
    setSelectedStyles([...styles])
    if (currentCategory) {
      beersApi.getByCategory(currentCategory.id).then(setFilteredBeers)
    } else {
      setFilteredBeers(beers)
    }
  }

  if (loading) return <div style={{ padding: 40, textAlign: 'center' }}>Загрузка...</div>

  return (
      <div className="catalog-layout">
        <aside className="sidebar">
          <div className="sidebar-section">
            <div className="sidebar-title">Категории</div>
            {categoryChildren.map(cat => (
                <div key={cat.id} className="tree-item" onClick={() => selectCategory(cat)} style={{ cursor: 'pointer' }}>
                  {cat.name}
                  <span className="tree-cnt">{beers.filter(b => b.category?.name === cat.name).length}</span>
                </div>
            ))}
            {categoryPath.length > 0 && (
                <button onClick={goBack} style={{ marginTop: 12, width: '100%', padding: 8, background: '#FAE8DF', border: 'none', borderRadius: 6, cursor: 'pointer' }}>
                  ← Назад
                </button>
            )}
          </div>
          <div className="divider" />

          <div className="sidebar-section">
            <div className="sidebar-title">Крепость (% ABV)</div>
            <div className="range-wrapper">
              <div className="range-row"><input type="range" min="0" max="20" value={filters.abvMin} step="0.5" onChange={e => setFilters({ ...filters, abvMin: +e.target.value })} /><span className="range-val">{filters.abvMin}%</span></div>
              <div className="range-row"><input type="range" min="0" max="20" value={filters.abvMax} step="0.5" onChange={e => setFilters({ ...filters, abvMax: +e.target.value })} /><span className="range-val">{filters.abvMax}%</span></div>
              <div className="range-bounds"><span>0%</span><span>20%</span></div>
            </div>
          </div>

          <div className="sidebar-section">
            <div className="sidebar-title">Горечь (IBU)</div>
            <div className="range-wrapper">
              <div className="range-row"><input type="range" min="0" max="120" value={filters.ibuMin} onChange={e => setFilters({ ...filters, ibuMin: +e.target.value })} /><span className="range-val">{filters.ibuMin}</span></div>
              <div className="range-row"><input type="range" min="0" max="120" value={filters.ibuMax} onChange={e => setFilters({ ...filters, ibuMax: +e.target.value })} /><span className="range-val">{filters.ibuMax}</span></div>
              <div className="range-bounds"><span>0</span><span>120</span></div>
            </div>
          </div>

          {countries.length > 0 && (
              <div className="sidebar-section">
                <div className="sidebar-title">Страна</div>
                {countries.map(c => (
                    <label key={c} className="check-row">
                      <input type="checkbox" checked={selectedCountries.includes(c)} onChange={() => setSelectedCountries(prev => prev.includes(c) ? prev.filter(x => x !== c) : [...prev, c])} />
                      {c}
                    </label>
                ))}
              </div>
          )}

          {styles.length > 0 && (
              <div className="sidebar-section">
                <div className="sidebar-title">Стиль</div>
                {styles.map(s => (
                    <label key={s} className="check-row">
                      <input type="checkbox" checked={selectedStyles.includes(s)} onChange={() => setSelectedStyles(prev => prev.includes(s) ? prev.filter(x => x !== s) : [...prev, s])} />
                      {s}
                    </label>
                ))}
              </div>
          )}

          <button className="search-btn" onClick={applyFilters}>🔍 Найти пиво</button>
          <button className="reset-filters" onClick={resetFilters}>⟳ Сбросить фильтры</button>
        </aside>

        <div>
          <div className="content-header">
            <div className="breadcrumb">
              <a onClick={goToRoot} style={{ cursor: 'pointer' }}>Каталог</a>
              {categoryPath.map((c, i) => (
                  <span key={c.id}>
                <span className="sep">›</span>
                <a onClick={() => selectCategory(c)} style={{ cursor: 'pointer' }}>{c.name}</a>
              </span>
              ))}
            </div>
            <span className="result-count">{filteredBeers.length} сортов</span>
          </div>

          <div className="beer-grid">
            {filteredBeers.length === 0 ? (
                <div className="no-results">🍺 Ничего не найдено<br />Попробуйте изменить параметры</div>
            ) : (
                filteredBeers.map(beer => (
                    <Link key={beer.id} to={`/beer/${beer.id}`} className="beer-card">
                      <div className="beer-thumb">🍺</div>
                      <div className="beer-name">{beer.name}</div>
                      <div className="beer-sub">{beer.country} · {beer.abv}% · {beer.ibu} IBU</div>
                      <div className="beer-tag">{beer.category?.name}</div>
                    </Link>
                ))
            )}
          </div>
        </div>
      </div>
  )
}