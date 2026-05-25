import { useState, useEffect } from 'react'
import { auth } from '../api/client'

export function useAuth() {
    const [isAuthenticated, setIsAuthenticated] = useState(false)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        setIsAuthenticated(auth.isAuthenticated())
        setLoading(false)
    }, [])

    const login = async (password) => {
        await auth.login(password)
        setIsAuthenticated(true)
    }

    const logout = () => {
        auth.logout()
        setIsAuthenticated(false)
    }

    return { isAuthenticated, loading, login, logout }
}