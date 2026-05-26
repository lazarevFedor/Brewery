import { Outlet, NavLink, Navigate } from 'react-router-dom'

const tabs = [
    { to: '/admin/beers', label: 'Пиво' },
    { to: '/admin/params', label: 'Параметры' },
    { to: '/admin/aggregates', label: 'Агрегаты' },
    { to: '/admin/categories', label: 'Категории' },
    { to: '/admin/enums', label: 'Перечисления' },
]

export function AdminPanel() {
    return (
        <div>
            <div style={{ display: 'flex', gap: '4px', marginBottom: '20px', borderBottom: '1px solid var(--border)', paddingBottom: '0' }}>
                {tabs.map(t => (
                    <NavLink
                        key={t.to}
                        to={t.to}
                        style={({ isActive }) => ({
                            padding: '8px 16px',
                            fontSize: '13px',
                            fontWeight: 600,
                            textDecoration: 'none',
                            borderRadius: 'var(--radius-sm) var(--radius-sm) 0 0',
                            color: isActive ? 'var(--accent)' : 'var(--text-muted)',
                            background: isActive ? 'var(--accent-light)' : 'transparent',
                            borderBottom: isActive ? '2px solid var(--accent)' : '2px solid transparent',
                            transition: 'all .15s',
                        })}
                    >
                        {t.label}
                    </NavLink>
                ))}
            </div>
            <Outlet />
        </div>
    )
}
