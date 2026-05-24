import { createContext, useContext, useState, useCallback } from 'react'

const ToastContext = createContext()

export function ToastProvider({ children }) {
    const [toast, setToast] = useState({ message: '', type: 'info', visible: false })

    const showToast = useCallback((message, type = 'info') => {
        setToast({ message, type, visible: true })
        setTimeout(() => setToast(prev => ({ ...prev, visible: false })), 3000)
    }, [])

    return (
        <ToastContext.Provider value={{ toast, showToast }}>
            {children}
        </ToastContext.Provider>
    )
}

export function useToast() {
    const context = useContext(ToastContext)
    if (!context) throw new Error('useToast must be used within ToastProvider')
    return context
}