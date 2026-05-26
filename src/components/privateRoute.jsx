import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

export function PrivateRoute() {
    const { isAuthenticated, loading } = useAuth()

    if (loading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Загрузка...</div>

    return isAuthenticated ? <Outlet /> : <Navigate to="/login" replace />
}