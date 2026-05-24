import { useState, useEffect } from 'react'
import { categoriesApi } from '../api/client'
import { useToast } from '../hooks/useToast'
import { Modal } from '../components/Modal'

export function Categories() {
    const [categories, setCategories] = useState([])
    const [modalOpen, setModalOpen] = useState(false)
    const [editing, setEditing] = useState(null)
    const [formData, setFormData] = useState({ name: '', parent_id: null })
    const { showToast } = useToast()

    useEffect(() => {
        loadCategories()
    }, [])

    const loadCategories = async () => {
        try {
            const data = await categoriesApi.getAll()
            setCategories(data)
        } catch {
            showToast('Ошибка загрузки', 'error')
        }
    }

    const handleSave = async () => {
        if (!formData.name.trim()) {
            showToast('Введите название', 'error')
            return
        }
        try {
            if (editing) {
                await categoriesApi.update(editing.id, { name: formData.name, parent_id: formData.parent_id })
                showToast('Категория обновлена', 'success')
            } else {
                await categoriesApi.create({ name: formData.name, parent_id: formData.parent_id })
                showToast('Категория создана', 'success')
            }
            setModalOpen(false)
            setEditing(null)
            setFormData({ name: '', parent_id: null })
            loadCategories()
        } catch {
            showToast('Ошибка сохранения', 'error')
        }
    }

    const handleDelete = async (id) => {
        if (!confirm('Удалить категорию?')) return
        try {
            await categoriesApi.delete(id)
            showToast('Категория удалена', 'success')
            loadCategories()
        } catch {
            showToast('Ошибка удаления', 'error')
        }
    }

    const getParentName = (parentId) => {
        if (!parentId) return '—'
        const parent = categories.find(c => c.id === parentId)
        return parent ? parent.name : '—'
    }

    return (
        <div>
            <div className="admin-header">
                <div>
                    <div className="admin-title">Категории товаров</div>
                    <div className="admin-sub">Управление категориями пива</div>
                </div>
                <button className="btn btn-primary" onClick={() => { setEditing(null); setFormData({ name: '', parent_id: null }); setModalOpen(true) }}>+ Добавить категорию</button>
            </div>

            <div className="table-wrap">
                <table>
                    <thead><tr><th>ID</th><th>Название</th><th>Родитель</th><th></th></tr></thead>
                    <tbody>
                    {categories.map(cat => (
                        <tr key={cat.id}>
                            <td style={{ color: 'var(--text-hint)' }}>{cat.id}</td>
                            <td><strong>{cat.name}</strong></td>
                            <td style={{ color: 'var(--text-muted)' }}>{getParentName(cat.parent_id)}</td>
                            <td>
                                <button className="icon-btn" onClick={() => { setEditing(cat); setFormData({ name: cat.name, parent_id: cat.parent_id }); setModalOpen(true) }} style={{ marginRight: '6px' }}>✏️</button>
                                <button className="icon-btn del" onClick={() => handleDelete(cat.id)}>🗑</button>
                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>

            <Modal isOpen={modalOpen} onClose={() => { setModalOpen(false); setEditing(null) }} title={editing ? 'Редактировать категорию' : 'Новая категория'}>
                <div className="form-row">
                    <label className="form-label">Название</label>
                    <input type="text" className="form-input" value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })} />
                </div>
                <div className="form-row">
                    <label className="form-label">Родительская категория</label>
                    <select className="form-select" value={formData.parent_id || ''} onChange={(e) => setFormData({ ...formData, parent_id: e.target.value ? Number(e.target.value) : null })}>
                        <option value="">— Без родителя —</option>
                        {categories.filter(c => c.id !== editing?.id).map(cat => <option key={cat.id} value={cat.id}>{cat.name}</option>)}
                    </select>
                </div>
                <div className="modal-footer">
                    <button className="btn btn-ghost" onClick={() => setModalOpen(false)}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleSave}>Сохранить</button>
                </div>
            </Modal>
        </div>
    )
}