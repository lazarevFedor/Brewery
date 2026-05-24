import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { useToast } from '../hooks/useToast'

export function Login() {
  const [password, setPassword] = useState('')
  const { login } = useAuth()
  const { showToast } = useToast()
  const navigate = useNavigate()

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      await login(password)
      showToast('Добро пожаловать!', 'success')
      navigate('/admin/params')
    } catch {
      showToast('Неверный пароль', 'error')
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
            <button type="submit" className="btn btn-primary btn-login">Войти</button>
          </form>
          <div className="admin-hint">🔑 Тестовые данные: <strong>admin123</strong></div>
        </div>
      </div>
  )
}