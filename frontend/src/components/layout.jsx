import { Outlet, Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth.jsx'
import '../styles/main.css'

export function Layout() {
    const { isAuthenticated, logout } = useAuth()
    const navigate = useNavigate()

    const handleAdminClick = () => {
        if (isAuthenticated) {
            navigate('/admin/params')
        } else {
            navigate('/login')
        }
    }

    const isActive = (path) => window.location.pathname === path ? 'active' : ''

    return (
        <div className="app">
            <nav className="navbar">
                <Link to="/" className="nav-logo">
                    🍺 ПивоКаталог
                </Link>
                <div className="nav-links">
                    <Link to="/catalog" className={`nav-link ${isActive('/catalog')}`}>Каталог</Link>
                    <Link to="/promotions" className={`nav-link ${isActive('/promotions')}`}>Акции</Link>
                    <Link to="/order-info" className={`nav-link ${isActive('/order-info')}`}>Как заказать</Link>
                    <Link to="/contacts" className={`nav-link ${isActive('/contacts')}`}>Контакты</Link>
                </div>
                <div style={{ display: 'flex', gap: '8px' }}>
                    <button
                        className={`admin-login-btn ${isAuthenticated ? 'admin-panel-btn' : ''}`}
                        onClick={handleAdminClick}
                    >
                        {isAuthenticated ? '⚙️ Админ-панель' : '🔐 Вход для админа'}
                    </button>
                    {isAuthenticated && (
                        <button
                            className="admin-login-btn"
                            onClick={() => { logout(); navigate('/') }}
                        >
                            Выйти
                        </button>
                    )}
                </div>
            </nav>
            <main className="main">
                <Outlet />
            </main>
        </div>
    )
}