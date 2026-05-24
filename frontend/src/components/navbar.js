// src/components/navbar.js
export function renderNavbar() {``
    return `
        <nav class="navbar">
            <a class="nav-logo" onclick="navigate('/')">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M17 11H6a4 4 0 0 0 0 8h11"/><path d="M17 11a4 4 0 0 1 0 8"/>
                    <path d="M10 3v5a2 2 0 0 0 2 2h0"/><path d="M14 3v5a2 2 0 0 0-2 2"/>
                </svg>
                Пивной животик
            </a>
            <div class="nav-links">
                <a class="nav-link active" onclick="navigate('/')>Каталог</a>
            </div>  
        </nav>
    `
}