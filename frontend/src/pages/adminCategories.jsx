import { useState, useEffect } from 'react'
import { categoriesApi } from '../api/client'
import { useToast } from '../hooks/useToast'
import { Modal } from '../components/Modal'
import '../styles/admin.css'

export function AdminCategories() {
    const [categories, setCategories] = useState([])
    const [path, setPath] = useState([])          // breadcrumb: [{id, name}]
    const [modalOpen, setModalOpen] = useState(false)
    const [newCategoryName, setNewCategoryName] = useState('')
    const [newParentId, setNewParentId] = useState(null)
    const { showToast } = useToast()

    useEffect(() => { loadCategories() }, [])

    const loadCategories = async () => {
        try {
            const data = await categoriesApi.getAll()
            setCategories(data)
        } catch {
            showToast('Ошибка загрузки категорий', 'error')
        }
    }

    const currentParentId = path.length > 0 ? path[path.length - 1].id : null

    const visibleChildren = categories.filter(c =>
        currentParentId === null ? !c.parent_id : c.parent_id === currentParentId
    )

    const enterCategory = (cat) => setPath([...path, { id: cat.id, name: cat.name }])
    const goToLevel = (idx) => setPath(path.slice(0, idx + 1))
    const goToRoot = () => setPath([])

    const handleCreate = async () => {
        if (!newCategoryName.trim()) return
        try {
            await categoriesApi.create({
                name: newCategoryName,
                parent_id: newParentId ?? (currentParentId ?? null),
            })
            showToast('Категория создана', 'success')
            setModalOpen(false)
            setNewCategoryName('')
            setNewParentId(null)
            loadCategories()
        } catch {
            showToast('Ошибка создания', 'error')
        }
    }

    const handleDelete = async (id, e) => {
        e.stopPropagation()
        if (!confirm('Удалить категорию?')) return
        try {
            await categoriesApi.delete(id)
            showToast('Категория удалена', 'success')
            loadCategories()
        } catch {
            showToast('Ошибка удаления', 'error')
        }
    }

    const childCount = (id) => categories.filter(c => c.parent_id === id).length

    return (
        <div>
            <div className="admin-header">
                <div>
                    <div className="admin-title">Категории</div>
                    <div className="admin-sub">{categories.length} категорий всего</div>
                </div>
                <button className="btn btn-primary" onClick={() => { setNewParentId(currentParentId); setModalOpen(true) }}>
                    + Новая категория
                </button>
            </div>

            {/* Breadcrumb */}
            <div className="breadcrumb" style={{ marginBottom: '12px' }}>
                <a style={{ cursor: 'pointer', color: 'var(--blue)' }} onClick={goToRoot}>Все категории</a>
                {path.map((p, i) => (
                    <span key={p.id} style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <span className="sep">›</span>
                        {i < path.length - 1
                            ? <a style={{ cursor: 'pointer', color: 'var(--blue)' }} onClick={() => goToLevel(i)}>{p.name}</a>
                            : <span style={{ fontWeight: 600 }}>{p.name}</span>
                        }
                    </span>
                ))}
            </div>

            {/* Category list */}
            <div className="table-wrap">
                <table>
                    <thead>
                        <tr>
                            <th>Название</th>
                            <th style={{ width: 80, textAlign: 'center' }}>Подкатегорий</th>
                            <th style={{ width: 40 }}></th>
                        </tr>
                    </thead>
                    <tbody>
                        {visibleChildren.length === 0 && (
                            <tr><td colSpan={3} style={{ textAlign: 'center', color: 'var(--text-hint)', padding: '24px' }}>
                                Нет категорий на этом уровне
                            </td></tr>
                        )}
                        {visibleChildren.map(cat => (
                            <tr key={cat.id} style={{ cursor: childCount(cat.id) > 0 ? 'pointer' : 'default' }}
                                onClick={() => childCount(cat.id) > 0 && enterCategory(cat)}>
                                <td>
                                    <span style={{ fontWeight: 500 }}>{cat.name}</span>
                                    {childCount(cat.id) > 0 && (
                                        <span style={{ marginLeft: 8, fontSize: 11, color: 'var(--blue)' }}>
                                            открыть подкатегории →
                                        </span>
                                    )}
                                </td>
                                <td style={{ textAlign: 'center' }}>
                                    {childCount(cat.id) > 0
                                        ? <span className="badge badge-numeric">{childCount(cat.id)}</span>
                                        : <span style={{ color: 'var(--text-hint)', fontSize: 12 }}>—</span>
                                    }
                                </td>
                                <td>
                                    <button className="icon-btn del" onClick={(e) => handleDelete(cat.id, e)}>🗑</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title="Новая категория">
                <div className="form-row">
                    <label className="form-label">Название</label>
                    <input type="text" className="form-input" value={newCategoryName}
                        autoFocus
                        onChange={(e) => setNewCategoryName(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleCreate()} />
                </div>
                <div className="form-row">
                    <label className="form-label">Родительская категория</label>
                    <select className="form-select" value={newParentId ?? ''} onChange={(e) => setNewParentId(e.target.value ? parseInt(e.target.value) : null)}>
                        <option value="">Корневая</option>
                        {categories.map(cat => (
                            <option key={cat.id} value={cat.id}>{cat.name}</option>
                        ))}
                    </select>
                </div>
                <div className="modal-footer">
                    <button className="btn btn-ghost" onClick={() => setModalOpen(false)}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleCreate}>Создать</button>
                </div>
            </Modal>
        </div>
    )
}
