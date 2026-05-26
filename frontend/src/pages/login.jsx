import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { auth } from '../api/client'
import { useToast } from '../hooks/useToast'
import '../styles/login.css'

export function Login() {
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const { showToast } = useToast()
    const navigate = useNavigate()

    const handleSubmit = async (e) => {
        e.preventDefault()
        try {
            await auth.login(username, password)
            showToast('Добро пожаловать в админ-панель!', 'success')
            navigate('/admin/params')
        } catch (err) {
            showToast(err.message || 'Неверный логин или пароль', 'error')
        }
    }

    return (
        <div className="login-container">
            <div className="login-card">
                <div className="login-icon">🔐</div>
                <div className="login-title">Вход в админ-панель</div>
                <div className="login-sub">Только для администраторов</div>
                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <div className="input-icon">
                            <span>👤</span>
                            <input
                                type="text"
                                className="form-input"
                                placeholder="Логин"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                            />
                        </div>
                    </div>
                    <div className="form-group">
                        <div className="input-icon">
                            <span>🔒</span>
                            <input
                                type="password"
                                className="form-input"
                                placeholder="Пароль"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                            />
                        </div>
                    </div>
                    <button type="submit" className="btn btn-primary btn-login">Войти в админ-панель</button>
                </form>
                <div className="admin-hint">
                    🔑 Тестовые данные: <strong>admin / admin123</strong>
                </div>
                <Link to="/catalog" className="back-link">← Вернуться в каталог</Link>
            </div>
        </div>
    )
}