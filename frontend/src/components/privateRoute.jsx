import { Navigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth.jsx'

export function PrivateRoute({ children }) {
    const { isAuthenticated, loading } = useAuth()

    if (loading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Загрузка...</div>

    return isAuthenticated ? children : <Navigate to="/login" replace />
}