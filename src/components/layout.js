import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import './layout.css'

export function Layout() {
    const { logout } = useAuth()
    const location = useLocation()
    const navigate = useNavigate()

    const handleLogout = () => {
        logout()
        navigate('/login')
    }

    const isActive = (path) => location.pathname === path

    return (
        <div className="app">
            <nav className="navbar">
                <Link to="/" className="nav-logo">📚 ПивоКаталог</Link>
                <div className="nav-links">
                    <Link to="/categories" className={`nav-link ${isActive('/categories') ? 'active' : ''}`}>Категории</Link>
                    <Link to="/parameters" className={`nav-link ${isActive('/parameters') ? 'active' : ''}`}>Параметры</Link>
                    <Link to="/aggregates" className={`nav-link ${isActive('/aggregates') ? 'active' : ''}`}>Агрегаты</Link>
                    <Link to="/enums" className={`nav-link ${isActive('/enums') ? 'active' : ''}`}>Enum</Link>
                </div>
                <button onClick={handleLogout} className="logout-btn">🚪 Выйти</button>
            </nav>
            <main className="main">
                <Outlet />
            </main>
        </div>
    )
}