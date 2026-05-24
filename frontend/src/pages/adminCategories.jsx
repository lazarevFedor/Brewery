import { useState, useEffect } from 'react'
import { categoriesApi, parametersApi, aggregatesApi } from '../api/client'
import { useToast } from '../hooks/useToast'
import { Modal } from '../components/Modal'
import './styles/admin.css'

export function AdminCategories() {
    const [categories, setCategories] = useState([])
    const [selectedCategory, setSelectedCategory] = useState(null)
    const [modalOpen, setModalOpen] = useState(false)
    const [newCategoryName, setNewCategoryName] = useState('')
    const [newParentId, setNewParentId] = useState(null)
    const { showToast } = useToast()

    useEffect(() => {
        loadCategories()
    }, [])

    const loadCategories = async () => {
        try {
            const data = await categoriesApi.getAll()
            setCategories(data)
        } catch {
            showToast('Ошибка загрузки категорий', 'error')
        }
    }

    const handleCreate = async () => {
        if (!newCategoryName.trim()) return
        try {
            await categoriesApi.create({
                name: newCategoryName,
                parent_id: newParentId || null,
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

    const renderTree = (parentId = null, level = 0) => {
        const children = categories.filter(c => c.parent_id === parentId)
        if (children.length === 0) return null
        return children.map(cat => (
            <div key={cat.id} style={{ marginLeft: `${level * 20}px` }} className="tree-item" onClick={() => setSelectedCategory(cat)}>
                {cat.name} <span className="tree-cnt">{categories.filter(c => c.parent_id === cat.id).length}</span>
                <button className="icon-btn del" onClick={(e) => { e.stopPropagation(); handleDelete(cat.id) }} style={{ marginLeft: 'auto' }}>🗑</button>
                {renderTree(cat.id, level + 1)}
            </div>
        ))
    }

    return (
        <div>
            <div className="admin-header">
                <div>
                    <div className="admin-title">Категории</div>
                    <div className="admin-sub">Управление иерархией категорий</div>
                </div>
                <button className="btn btn-primary" onClick={() => setModalOpen(true)}>+ Новая категория</button>
            </div>

            <div className="table-wrap" style={{ padding: '16px' }}>
                <div className="enum-list">
                    <div className="tree-item" onClick={() => setSelectedCategory({ id: null, name: 'Корень', parent_id: null })}>
                        📁 Корень
                    </div>
                    {renderTree(null)}
                </div>
            </div>

            {selectedCategory && (
                <div style={{ marginTop: '24px' }}>
                    <div className="admin-header">
                        <div>
                            <div className="admin-title">Параметры категории: {selectedCategory.name}</div>
                            <div className="admin-sub">Наследуемые параметры будут применены автоматически</div>
                        </div>
                    </div>
                    <div className="table-wrap">
                        <table>
                            <thead><tr><th>Параметр</th><th>Тип</th><th>Значение</th><th>Источник</th></tr></thead>
                            <tbody>
                            <tr><td colSpan="4" style={{ textAlign: 'center', color: '#999' }}>Выберите категорию для просмотра параметров</td></tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            )}

            <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title="Новая категория">
                <div className="form-row">
                    <label className="form-label">Название категории</label>
                    <input type="text" className="form-input" value={newCategoryName} onChange={(e) => setNewCategoryName(e.target.value)} />
                </div>
                <div className="form-row">
                    <label className="form-label">Родительская категория</label>
                    <select className="form-select" value={newParentId || ''} onChange={(e) => setNewParentId(e.target.value ? parseInt(e.target.value) : null)}>
                        <option value="">Корневая категория</option>
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