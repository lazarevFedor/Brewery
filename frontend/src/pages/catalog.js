import { getBeers, getCategories, getCategoryChildren } from '../api/beers.js'
import { navigate } from '../router.js'
import { renderLayout } from '../components/layout.js'


export async function renderCatalog(container) {
  // Сразу показываем скелетон
  // container.innerHTML = `
  //   <div style="display:grid;grid-template-columns:220px 1fr;gap:16px">
  //     <div style="background:#f5f5f5;border-radius:8px;padding:16px;height:400px"></div>
  //     <div style="background:#f5f5f5;border-radius:8px;padding:16px;height:300px"></div>
  //   </div>
  // `

  container.innerHTML = renderLayout({
    content: `
      <div class="catalog-layout">
        <aside class="sidebar">
          <div class="skeleton" style="height:400px"></div>
        </aside>
        <div>
          <div class="skeleton" style="height:300px"></div>
        </div>
      </div>
    `
  })

  try {
    // Грузим данные параллельно
    const [allCategories, initialBeers] = await Promise.all([
      getCategories(),
      getBeers({ limit: 50 })
    ])
    const rootCategory = allCategories.filter(c => !c.parent_id)[0]
    const categories = await getCategoryChildren(rootCategory.id);

    // Рендерим полную страницу
    const html = renderCatalogLayout(categories, initialBeers)
    container.innerHTML = html

    // Инициализируем логику
    initCatalogLogic(categories, initialBeers)

  } catch (err) {
    container.innerHTML = `<div style="padding:20px;color:red">❌ Ошибка: ${err.message}</div>`
  }
}

// ═══════════════════════════════════════════════════════════
// ГЛАВНЫЙ РЕНДЕР
// ═══════════════════════════════════════════════════════════

function renderCatalogLayout(categories, beers) {
  return ` 
    <div style="
      display: grid;
      grid-template-columns: 220px 1fr;
      gap: 16px;
      padding: 0;
    ">
      <!-- SIDEBAR -->
      <aside style="
        background: white;
        border: 1px solid #e0e0e0;
        border-radius: 8px;
        padding: 16px;
        position: sticky;
        top: 72px;
        max-height: calc(100vh - 88px);
        overflow-y: auto;
      ">
        ${renderSidebar(categories, beers)}
      </aside>

      <!-- CONTENT -->
      <div>
        <div style="
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 16px;
          padding: 0 12px;
        ">
          <div style="
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 13px;
          ">
            <a onclick="window.catalog_navigate('/')" style="
              color: #0066cc;
              cursor: pointer;
              text-decoration: none;
            ">Все</a>
            <span style="color: #999">›</span>
            <span id="cat-label" style="font-weight: 600; color: #333">Все сорта</span>
          </div>
          <span id="result-count" style="
            font-size: 12px;
            color: #666;
          ">${beers.length} сортов</span>
        </div>

        <!-- BEER GRID -->
        <div id="beer-grid" style="
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
          gap: 10px;
        ">
          ${renderBeerCards(beers)}
        </div>
      </div>
    </div>
  `
}

function renderSidebar(categories, beers) {
  return `
    <!-- КАТЕГОРИИ -->
    <div style="margin-bottom: 20px;">
      <div style="
        font-size: 10px;
        color: #999;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        margin-bottom: 8px;
        font-weight: 600;
      ">Категории</div>

      <div class="tree-item-root" data-id="root" style="
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 6px 8px;
        border-radius: 6px;
        font-size: 13px;
        color: #666;
        cursor: pointer;
        background: #e8f0ff;
        color: #0066cc;
        font-weight: 600;
        margin-bottom: 4px;
      " onclick="window.catalog_selectCat('root', 'Все сорта')">
        <span>Все сорта</span>
        <span style="font-size:10px;background:rgba(0,102,204,0.15);padding:1px 7px;border-radius:20px">${categories.length || 0}</span>
      </div>

      ${renderCategoryTree(categories, beers)}
    </div>

    <div style="height:1px;background:#e0e0e0;margin:16px 0"></div>

    <!-- ФИЛЬТР: ABV -->
    <div style="margin-bottom: 20px;">
      <div style="
        font-size: 10px;
        color: #999;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        margin-bottom: 8px;
        font-weight: 600;
      ">Крепость (% ABV)</div>

      <div style="margin-bottom: 12px;">
        <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px">
          <input type="range" min="0" max="20" value="0" id="abv-min" 
            class="filter-input" step="0.5"
            style="flex:1;accent-color:#0066cc;height:3px;cursor:pointer">
          <span id="abv-min-val" style="
            font-size:12px;
            font-weight:600;
            color:#0066cc;
            min-width:32px;
            text-align:right
          ">0%</span>
        </div>
        <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px">
          <input type="range" min="0" max="20" value="20" id="abv-max"
            class="filter-input" step="0.5"
            style="flex:1;accent-color:#0066cc;height:3px;cursor:pointer">
          <span id="abv-max-val" style="
            font-size:12px;
            font-weight:600;
            color:#0066cc;
            min-width:32px;
            text-align:right
          ">20%</span>
        </div>
        <div style="display:flex;justify-content:space-between;font-size:10px;color:#999;margin-top:2px">
          <span>0%</span><span>20%</span>
        </div>
      </div>
    </div>

    <!-- ФИЛЬТР: IBU -->
    <div style="margin-bottom: 20px;">
      <div style="
        font-size: 10px;
        color: #999;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        margin-bottom: 8px;
        font-weight: 600;
      ">Горечь (IBU)</div>

      <div style="margin-bottom: 12px;">
        <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px">
          <input type="range" min="0" max="120" value="0" id="ibu-min"
            class="filter-input"
            style="flex:1;accent-color:#0066cc;height:3px;cursor:pointer">
          <span id="ibu-min-val" style="
            font-size:12px;
            font-weight:600;
            color:#0066cc;
            min-width:32px;
            text-align:right
          ">0</span>
        </div>
        <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px">
          <input type="range" min="0" max="120" value="120" id="ibu-max"
            class="filter-input"
            style="flex:1;accent-color:#0066cc;height:3px;cursor:pointer">
          <span id="ibu-max-val" style="
            font-size:12px;
            font-weight:600;
            color:#0066cc;
            min-width:32px;
            text-align:right
          ">120</span>
        </div>
        <div style="display:flex;justify-content:space-between;font-size:10px;color:#999;margin-top:2px">
          <span>0</span><span>120</span>
        </div>
      </div>
    </div>

    <!-- ФИЛЬТР: Страна -->
    <div style="margin-bottom: 20px;">
      <div style="
        font-size: 10px;
        color: #999;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        margin-bottom: 8px;
        font-weight: 600;
      ">Страна</div>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="Бельгия" class="country-filter filter-input" checked
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>Бельгия</span>
      </label>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="Германия" class="country-filter filter-input" checked
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>Германия</span>
      </label>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="США" class="country-filter filter-input"
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>США</span>
      </label>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="Чехия" class="country-filter filter-input"
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>Чехия</span>
      </label>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="Россия" class="country-filter filter-input"
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>Россия</span>
      </label>
    </div>

    <!-- ФИЛЬТР: Стиль -->
    <div style="margin-bottom: 20px;">
      <div style="
        font-size: 10px;
        color: #999;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        margin-bottom: 8px;
        font-weight: 600;
      ">Стиль</div>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="IPA" class="style-filter filter-input" checked
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>IPA</span>
      </label>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="Пейл эль" class="style-filter filter-input"
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>Пейл эль</span>
      </label>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="Сэзон" class="style-filter filter-input"
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>Сэзон</span>
      </label>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="Трапист" class="style-filter filter-input"
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>Трапист</span>
      </label>
      <label style="display:flex;align-items:center;gap:8px;font-size:13px;padding:3px 0;cursor:pointer">
        <input type="checkbox" value="Стаут" class="style-filter filter-input"
          style="width:14px;height:14px;accent-color:#0066cc;cursor:pointer">
        <span>Стаут</span>
      </label>
    </div>

    <!-- КНОПКИ -->
    <button id="apply-filters-btn" style="
      display: block;
      width: 100%;
      padding: 10px;
      margin-bottom: 8px;
      border: none;
      border-radius: 6px;
      background: #0066cc;
      color: white;
      font-size: 13px;
      font-weight: 600;
      cursor: pointer;
    ">🔍 Найти пиво</button>

    <button id="reset-filters-btn" style="
      display: block;
      width: 100%;
      padding: 10px;
      border: 1px solid #e0e0e0;
      border-radius: 6px;
      background: white;
      color: #666;
      font-size: 13px;
      font-weight: 600;
      cursor: pointer;
    ">⟳ Сбросить</button>

    <div id="filter-summary" style="
      margin-top: 12px;
      font-size: 11px;
      color: #0066cc;
      text-align: center;
    "></div>
  `
}

function renderCategoryTree(categories, beers) {
  if (!categories || !Array.isArray(categories)) return ''

  // // 1. Быстро считаем количество пива для каждой категории (Группировка)
  // const beerCounts = beers.reduce((acc, beer) => {
  //   if (beer.category.name) {
  //     acc[beer.category.name] = (acc[beer.category.name] || 0) + 1;
  //   }
  //   return acc;
  // }, {});


  // ==========================================
  // 1. ПОДГОТОВИТЕЛЬНЫЕ РАСЧЕТЫ (В САМОМ НАЧАЛЕ)
  // ==========================================

  // Шаг A: Создаем карту "Имя категории -> Идентификатор ID"
  // Это нужно, так как в пиве есть только имя категории (beer.category.name)
  const categoryIdByName = categories.reduce((acc, cat) => {
    if (cat.name) {
      acc[cat.name.trim().toLowerCase()] = String(cat.id);
    }
    return acc;
  }, {});

  // Шаг B: Считаем пиво напрямую привязанное к каждой категории по имени
  const directBeerCounts = beers.reduce((acc, beer) => {
    const catName = beer.category?.name;
    if (catName) {
      const catId = categoryIdByName[catName.trim().toLowerCase()];
      if (catId) {
        acc[catId] = (acc[catId] || 0) + 1;
      }
    }
    return acc;
  }, {});

  // Шаг C: Строим карту связей "Родитель -> Список ID детей"
  const childrenMap = categories.reduce((acc, cat) => {
    // Если parent_id равен 0, null или undefined — это корень
    const parentId = cat.parent_id ? String(cat.parent_id) : 'root';
    if (!acc[parentId]) acc[parentId] = [];
    acc[parentId].push(String(cat.id));
    return acc;
  }, {});

  // Шаг D: Рекурсивный расчет сумм с кэшированием (мемоизацией)
  const totalCountsCache = {};

  function calculateTotalCount(categoryId) {
    const id = categoryId;
    
    if (id in totalCountsCache) return totalCountsCache[id];

    // Берем прямое пиво
    let sum = directBeerCounts[id] || 0;

    // Добавляем пиво из всех подкатегорий
    const children = childrenMap[id] || [];
    for (const childId of children) {
      sum += calculateTotalCount(childId);
    }

    totalCountsCache[id] = sum;
    return sum;
  }

  // Шаг E: Запускаем расчет для ВСЕХ категорий заранее, чтобы наполнить кэш
  categories.forEach(cat => calculateTotalCount(cat.id));

  console.log(beers)

  return categories.map(c => {
    const count = totalCountsCache[c.id] || 0;
    return `   
    <div class="tree-item" data-id="${c.id}"
    onmouseover="this.style.background='#f5f5f5'" onmouseout="this.style.background='transparent'"
      onclick="window.catalog_selectCat('${c.id}', '${c.name}')">
      <span>${escapeHtml(c.name)}</span>
      <span style="font-size:10px;background:#f5f5f5;padding:1px 7px;border-radius:20px;color:#999">${count}</span>
    </div>
  `;
  }).join('');
}

function renderBeerCards(beers) {
  if (!beers || beers.length === 0) {
    return `<div style="
      grid-column: 1 / -1;
      padding: 40px;
      text-align: center;
      color: #999;
    ">🍺 Ничего не найдено</div>`
  }

  return beers.map(beer => `
    <a href="#" onclick="window.catalog_navigate('/beer/${beer.id}');return false" style="
      display: flex;
      flex-direction: column;
      background: white;
      border: 1px solid #e0e0e0;
      border-radius: 8px;
      padding: 12px;
      cursor: pointer;
      text-decoration: none;
      transition: all 0.15s;
    " onmouseover="this.style.borderColor='#0066cc';this.style.transform='translateY(-2px)'"
      onmouseout="this.style.borderColor='#e0e0e0';this.style.transform='translateY(0)'">
      <div style="
        width: 100%;
        height: 80px;
        background: #f9f9f9;
        border-radius: 6px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 32px;
        margin-bottom: 10px;
      ">${beer.emoji || '🍺'}</div>
      <div style="
        font-size: 13px;
        font-weight: 600;
        color: #333;
        margin-bottom: 3px;
        line-height: 1.3;
      ">${escapeHtml(beer.name)}</div>
      <div style="
        font-size: 11px;
        color: #999;
        margin-bottom: 8px;
      ">${escapeHtml(beer.country)} · ${beer.abv}% · ${beer.ibu} IBU</div>
      <div style="
        display: inline-block;
        padding: 2px 8px;
        border-radius: 20px;
        font-size: 10px;
        font-weight: 600;
        background: #fff5f0;
        color: #c4541a;
      ">${escapeHtml(beer.style)}</div>
    </a>
  `).join('')
}

// ═══════════════════════════════════════════════════════════
// ЛОГИКА И СОБЫТИЯ
// ═══════════════════════════════════════════════════════════

function initCatalogLogic(allCategories, allBeers) {
  window.catalog_data = {
    categories: allCategories,
    beers: allBeers,
    currentCategory: null
  }

  // События на фильтры
  document.querySelectorAll('.filter-input').forEach(input => {
    input.addEventListener('input', updateRangeDisplay)
    input.addEventListener('change', applyFilters)
  })

  // Кнопки
  document.getElementById('apply-filters-btn')?.addEventListener('click', applyFilters)
  document.getElementById('reset-filters-btn')?.addEventListener('click', resetFilters)

  // Экспортируем глобальные функции
  window.catalog_selectCat = selectCategory
  window.catalog_navigate = navigate
  window.catalog_applyFilters = applyFilters
}

function getFilterParams() {
  return {
    abvMin: parseFloat(document.getElementById('abv-min')?.value ?? 0),
    abvMax: parseFloat(document.getElementById('abv-max')?.value ?? 20),
    ibuMin: parseInt(document.getElementById('ibu-min')?.value ?? 0),
    ibuMax: parseInt(document.getElementById('ibu-max')?.value ?? 120),
    countries: Array.from(document.querySelectorAll('.country-filter:checked')).map(cb => cb.value),
    styles: Array.from(document.querySelectorAll('.style-filter:checked')).map(cb => cb.value)
  }
}

function filterBeers(beers, params) {
  return beers.filter(beer => {
    const abv = parseFloat(beer.abv?.toString() ?? 0)
    const ibu = parseInt(beer.ibu?.toString() ?? 0)
    const country = beer.country
    const style = beer.style

    if (abv < params.abvMin || abv > params.abvMax) return false
    if (ibu < params.ibuMin || ibu > params.ibuMax) return false
    if (params.countries.length > 0 && !params.countries.includes(country)) return false
    if (params.styles.length > 0 && !params.styles.includes(style)) return false

    return true
  })
}

function applyFilters() {
  const data = window.catalog_data
  if (!data) return

  const params = getFilterParams()
  const filtered = filterBeers(data.beers, params)

  const grid = document.getElementById('beer-grid')
  if (grid) {
    grid.innerHTML = renderBeerCards(filtered)
  }

  const count = document.getElementById('result-count')
  if (count) {
    count.textContent = `${filtered.length} ${getWordForm(filtered.length, ['сорт', 'сорта', 'сортов'])}`
  }

  updateFilterSummary(params)
}

function updateRangeDisplay(e) {
  if (e.target.id === 'abv-min') document.getElementById('abv-min-val').textContent = e.target.value + '%'
  if (e.target.id === 'abv-max') document.getElementById('abv-max-val').textContent = e.target.value + '%'
  if (e.target.id === 'ibu-min') document.getElementById('ibu-min-val').textContent = e.target.value
  if (e.target.id === 'ibu-max') document.getElementById('ibu-max-val').textContent = e.target.value
}

function updateFilterSummary(params) {
  const parts = []
  if (params.abvMin > 0 || params.abvMax < 20) {
    parts.push(`крепость ${params.abvMin}–${params.abvMax}%`)
  }
  if (params.ibuMin > 0 || params.ibuMax < 120) {
    parts.push(`горечь ${params.ibuMin}–${params.ibuMax}`)
  }
  if (params.countries.length > 0 && params.countries.length < 5) {
    parts.push(`${params.countries.join(', ')}`)
  }
  if (params.styles.length > 0 && params.styles.length < 5) {
    parts.push(`стили: ${params.styles.join(', ')}`)
  }

  const el = document.getElementById('filter-summary')
  if (el) {
    el.textContent = parts.length > 0 ? `⚡ ${parts.join(' · ')}` : ''
  }
}

function resetFilters() {
  document.getElementById('abv-min').value = 0
  document.getElementById('abv-max').value = 20
  document.getElementById('ibu-min').value = 0
  document.getElementById('ibu-max').value = 120

  document.querySelectorAll('.country-filter').forEach(cb => {
    cb.checked = cb.value === 'Бельгия' || cb.value === 'Германия'
  })

  document.querySelectorAll('.style-filter').forEach(cb => {
    cb.checked = true
  })

  document.querySelectorAll('.tree-item').forEach(item => item.classList.remove('active'))
  document.querySelector('.tree-item-root')?.style.setProperty('background', '#e8f0ff')
  document.getElementById('cat-label').textContent = 'Все сорта'

  updateRangeDisplay({ target: { id: 'abv-min' } })
  updateRangeDisplay({ target: { id: 'abv-max' } })
  updateRangeDisplay({ target: { id: 'ibu-min' } })
  updateRangeDisplay({ target: { id: 'ibu-max' } })

  applyFilters()
}

function selectCategory(catId, catName) {
  window.catalog_data.currentCategory = catId

  document.querySelectorAll('.tree-item').forEach(item => item.style.background = 'transparent')
  document.querySelector(`[data-id="${catId}"]`)?.style.setProperty('background', '#e8f0ff')
  document.querySelector(`[data-id="${catId}"]`)?.style.setProperty('color', '#0066cc')
  document.querySelector(`[data-id="${catId}"]`)?.style.setProperty('font-weight', '600')
  
  document.getElementById('cat-label').textContent = catName

  if (catId !== 'root') {
    document.querySelectorAll('.style-filter').forEach(cb => cb.checked = false)
    const styleCheckbox = Array.from(document.querySelectorAll('.style-filter')).find(cb => cb.value === catName)
    if (styleCheckbox) styleCheckbox.checked = true
  } else {
    document.querySelectorAll('.style-filter').forEach(cb => cb.checked = true)
  }

  applyFilters()
}

function getWordForm(count, forms) {
  if (count % 10 === 1 && count % 100 !== 11) return forms[0]
  if ([2, 3, 4].includes(count % 10) && ![12, 13, 14].includes(count % 100)) return forms[1]
  return forms[2]
}

function escapeHtml(text) {
  const map = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;' }
  return text?.replace(/[&<>"']/g, m => map[m]) ?? ''
}



// function countCategoryAmount(categories, beers) {
//   // ==========================================
//   // 1. ПОДГОТОВИТЕЛЬНЫЕ РАСЧЕТЫ (В САМОМ НАЧАЛЕ)
//   // ==========================================

//   // Шаг A: Создаем карту "Имя категории -> Идентификатор ID"
//   // Это нужно, так как в пиве есть только имя категории (beer.category.name)
//   const categoryIdByName = categories.reduce((acc, cat) => {
//     if (cat.name) {
//       acc[cat.name] = cat.id;
//     }
//     return acc;
//   }, {});

//   // Шаг B: Считаем пиво напрямую привязанное к каждой категории по имени
//   const directBeerCounts = beers.reduce((acc, beer) => {
//     const catName = beer.category?.name;
//     if (catName) {
//       const catId = categoryIdByName[catName];
//       if (catId) {
//         acc[catId] = (acc[catId] || 0) + 1;
//       }
//     }
//     return acc;
//   }, {});

//   // Шаг C: Строим карту связей "Родитель -> Список ID детей"
//   const childrenMap = categories.reduce((acc, cat) => {
//     // Если parent_id равен 0, null или undefined — это корень
//     const parentId = cat.parent_id ? String(cat.parent_id) : 'root';
//     if (!acc[parentId]) acc[parentId] = [];
//     acc[parentId].push(String(cat.id));
//     return acc;
//   }, {});

//   // Шаг D: Рекурсивный расчет сумм с кэшированием (мемоизацией)
//   const totalCountsCache = {};

//   function calculateTotalCount(categoryId) {
//     const idStr = String(categoryId);
    
//     if (idStr in totalCountsCache) return totalCountsCache[idStr];

//     // Берем прямое пиво
//     let sum = directBeerCounts[idStr] || 0;

//     // Добавляем пиво из всех подкатегорий
//     const children = childrenMap[idStr] || [];
//     for (const childId of children) {
//       sum += calculateTotalCount(childId);
//     }

//     totalCountsCache[idStr] = sum;
//     return sum;
//   }

//   // Шаг E: Запускаем расчет для ВСЕХ категорий заранее, чтобы наполнить кэш
//   categories.forEach(cat => calculateTotalCount(cat.id));

// }
