import { useToast } from '../hooks/useToast'

export function Toast() {
    const { toast } = useToast()

    if (!toast.visible) return null

    return <div className={`toast show ${toast.type}`}>{toast.message}</div>
}