import { renderNavbar } from './navbar.js'

// Оборачивает любую страницу в общий шаблон
export function renderLayout(content) {
    return `
        <header>
            ${renderNavbar()}
        </header>
        <main class="main">
            ${content}
        </main>
    `
}