import { useState, useEffect } from 'react'

export function useAuth() {
    const [isAuthenticated, setIsAuthenticated] = useState(false)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        const token = localStorage.getItem('admin_token')
        setIsAuthenticated(!!token)
        setLoading(false)
    }, [])

    const login = (password) => {
        if (password === 'admin123') {
            localStorage.setItem('admin_token', 'dummy-token')
            setIsAuthenticated(true)
            return Promise.resolve(true)
        }
        return Promise.reject(new Error('Invalid password'))
    }

    const logout = () => {
        localStorage.removeItem('admin_token')
        setIsAuthenticated(false)
    }

    return { isAuthenticated, loading, login, logout }
}