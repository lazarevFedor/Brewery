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
            navigate('/')
        } catch (err) {
            showToast(err.message, 'error')
        }
    }

    return (
        <div style={{ minHeight: 'calc(100vh - 56px)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '2rem' }}>
            <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: '16px', padding: '2rem', maxWidth: '420px', width: '100%', boxShadow: 'var(--shadow-md)' }}>
                <div style={{ fontSize: '56px', textAlign: 'center', marginBottom: '1rem' }}>🔐</div>
                <div style={{ fontFamily: "'Unbounded', sans-serif", fontSize: '20px', fontWeight: 700, textAlign: 'center', marginBottom: '0.5rem' }}>Вход в админ-панель</div>
                <div style={{ fontSize: '13px', color: 'var(--text-muted)', textAlign: 'center', marginBottom: '1.5rem' }}>Только для администраторов</div>
                <form onSubmit={handleSubmit}>
                    <div style={{ marginBottom: '1.25rem' }}>
                        <div style={{ position: 'relative' }}>
                            <span style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', fontSize: '18px', color: '#999' }}>🔒</span>
                            <input type="password" className="form-input" style={{ paddingLeft: '40px' }} placeholder="Пароль" value={password} onChange={(e) => setPassword(e.target.value)} />
                        </div>
                    </div>
                    <button type="submit" className="btn btn-primary" style={{ width: '100%', padding: '12px' }}>Войти в админ-панель</button>
                </form>
                <div style={{ marginTop: '1.5rem', paddingTop: '1rem', borderTop: '1px solid var(--border)', fontSize: '12px', color: '#999', textAlign: 'center' }}>🔑 Тестовые данные: <strong>admin123</strong></div>
            </div>
        </div>
    )
}