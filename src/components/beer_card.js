// src/components/beerCard.js
// Переиспользуемая карточка пива для каталога и поиска

export function renderBeerCard(beer) {
  return `
    <a href="#" class="beer-card" onclick="navigate('/beer/${beer.id}'); return false;">
      <div class="beer-thumb">${beer.emoji || '🍺'}</div>
      <div class="beer-name">${escapeHtml(beer.name)}</div>
      <div class="beer-sub">${escapeHtml(beer.country)} · ${beer.abv}% · ${beer.ibu} IBU</div>
      <div class="beer-tag">${escapeHtml(beer.style)}</div>
    </a>
  `
}

function escapeHtml(text) {
  const map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  }
  return text?.replace(/[&<>"']/g, m => map[m]) ?? ''
}