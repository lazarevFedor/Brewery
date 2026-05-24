// src/pages/detail.js
import { getBeer } from '../api/beers.js'
import { navigate } from '../router.js'

export async function renderDetail(container, params) {
    // params.id = '3'  ← пришло из роутера

    // Сразу показываем заглушку пока грузятся данные
    container.innerHTML = `
        <div class="detail-skeleton">
            <div class="skeleton" style="height:200px"></div>
            <div class="skeleton" style="height:40px;margin-top:12px"></div>
        </div>
    `

    // Запрашиваем данные — управление уходит в api/beers.js
    const beer = await getBeer(params.id)
    // ← здесь выполнение стоит, пока сервер не ответил
    // beer = { id: 3, name: 'Karmeliet', abv: 8.4, country: 'Бельгия', ... }

    // Теперь рендерим реальный HTML на основе данных с сервера
    container.innerHTML = `
        <div class="breadcrumb">
            <a onclick="navigate('/')">Все</a> › 
            <span>${beer.name}</span>
        </div>
        <div class="detail-grid">
            <div class="beer-hero">${beer.emoji ?? '🍺'}</div>
            <div>
                <h1 class="detail-title">${beer.name}</h1>
                <p class="detail-cat">${beer.category} · ${beer.country}</p>
                <div class="feat-grid">
                    <div class="feat-card">
                        <div class="feat-label">Крепость</div>
                        <div class="feat-val">${beer.abv}%</div>
                    </div>
                    <div class="feat-card">
                        <div class="feat-label">IBU</div>
                        <div class="feat-val">${beer.ibu}</div>
                    </div>
                </div>
            </div>
        </div>
    `

    // Вешаем события уже на реально существующие элементы
    bindDetailEvents(beer)
}

function bindDetailEvents(beer) {
    // Клики, формы и т.д.
}