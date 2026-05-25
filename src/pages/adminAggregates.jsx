import { useState, useEffect } from 'react'
import { aggregatesApi, parametersApi, categoriesApi } from '../api/client'
import { useToast } from '../hooks/useToast'
import { Modal } from '../components/Modal'
import './styles/admin.css'

export function AdminAggregates() {
    const [aggregates, setAggregates] = useState([])
    const [selectedAggregate, setSelectedAggregate] = useState(null)
    const [params, setParams] = useState({ numeric: [], enum: [] })
    const [categories, setCategories] = useState([])
    const [modalOpen, setModalOpen] = useState(false)
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        numeric_param_ids: [],
        enum_param_ids: [],
    })
    const { showToast } = useToast()

    useEffect(() => {
        loadAggregates()
        loadParams()
        loadCategories()
    }, [])

    const loadAggregates = async () => {
        try {
            const data = await aggregatesApi.getAll()
            setAggregates(data)
        } catch {
            showToast('Ошибка загрузки агрегатов', 'error')
        }
    }

    const loadParams = async () => {
        try {
            const numeric = await parametersApi.getAll(null, 'numeric')
            const enumParams = await parametersApi.getAll(null, 'enum')
            setParams({ numeric, enum: enumParams })
        } catch {
            // ignore
        }
    }

    const loadCategories = async () => {
        try {
            const data = await categoriesApi.getAll()
            setCategories(data)
        } catch {
            // ignore
        }
    }

    const handleCreate = async () => {
        if (!formData.name.trim()) {
            showToast('Введите название агрегата', 'error')
            return
        }
        try {
            await aggregatesApi.create(formData)
            showToast('Агрегат создан', 'success')
            setModalOpen(false)
            setFormData({ name: '', description: '', numeric_param_ids: [], enum_param_ids: [] })
            loadAggregates()
        } catch {
            showToast('Ошибка создания', 'error')
        }
    }

    const handleDelete = async (id) => {
        if (!confirm('Удалить агрегат?')) return
        try {
            await aggregatesApi.delete(id)
            showToast('Агрегат удалён', 'success')
            loadAggregates()
        } catch {
            showToast('Ошибка удаления', 'error')
        }
    }

    const handleApplyToCategory = async (aggregateId, categoryId) => {
        try {
            await aggregatesApi.applyToCategory(categoryId, aggregateId)
            showToast('Параметры применены к категории', 'success')
        } catch {
            showToast('Ошибка применения', 'error')
        }
    }

    return (
        <div>
            <div className="admin-header">
                <div>
                    <div className="admin-title">Агрегаты параметров</div>
                    <div className="admin-sub">Наборы параметров для быстрого применения к категориям</div>
                </div>
                <button className="btn btn-primary" onClick={() => setModalOpen(true)}>+ Новый агрегат</button>
            </div>

            <div className="enum-layout">
                <div className="enum-list">
                    {aggregates.map(agg => (
                        <div key={agg.id} className={`enum-class-item ${selectedAggregate?.id === agg.id ? 'active' : ''}`}
                             onClick={() => setSelectedAggregate(agg)}>
                            <span><strong>{agg.name}</strong><br /><small style={{ fontSize: '11px' }}>{agg.description}</small></span>
                            <div>
                                <span className="cnt">{agg.numeric_param_ids?.length || 0} числ.</span>
                                <span className="cnt" style={{ marginLeft: '4px' }}>{agg.enum_param_ids?.length || 0} enum</span>
                                <button className="icon-btn" onClick={(e) => { e.stopPropagation(); handleDelete(agg.id) }} style={{ marginLeft: '8px' }}>🗑</button>
                            </div>
                        </div>
                    ))}
                </div>

                <div>
                    {selectedAggregate && (
                        <>
                            <div className="admin-header">
                                <div>
                                    <div className="admin-title">{selectedAggregate.name}</div>
                                    <div className="admin-sub">{selectedAggregate.description || 'Нет описания'}</div>
                                </div>
                            </div>

                            <div className="table-wrap">
                                <table>
                                    <thead><tr><th>Параметр</th><th>Тип</th><th>Диапазон / Enum</th></tr></thead>
                                    <tbody>
                                    {selectedAggregate.numeric_param_ids?.map(id => {
                                        const p = params.numeric.find(p => p.id === id)
                                        return p ? (
                                            <tr key={p.id}><td>{p.field_name}</td><td><span className="badge badge-numeric">numeric</span></td><td>{p.min_val} — {p.max_val}</td></tr>
                                        ) : null
                                    })}
                                    {selectedAggregate.enum_param_ids?.map(id => {
                                        const p = params.enum.find(p => p.id === id)
                                        return p ? (
                                            <tr key={p.id}><td>Enum #{p.enum_class_id}</td><td><span className="badge badge-enum">enum</span></td><td>Enum class #{p.enum_class_id}</td></tr>
                                        ) : null
                                    })}
                                    </tbody>
                                </table>
                            </div>

                            <div className="admin-header" style={{ marginTop: '24px' }}>
                                <div className="admin-title">Применить к категории</div>
                            </div>
                            <div className="table-wrap">
                                <table>
                                    <thead><tr><th>Категория</th><th style={{ width: '120px' }}></th></tr></thead>
                                    <tbody>
                                    {categories.map(cat => (
                                        <tr key={cat.id}><td>{cat.name}</td>
                                            <td><button className="btn btn-primary" style={{ padding: '4px 12px', fontSize: '12px' }} onClick={() => handleApplyToCategory(selectedAggregate.id, cat.id)}>Применить</button></td>
                                        </tr>
                                    ))}
                                    </tbody>
                                </table>
                            </div>
                        </>
                    )}
                </div>
            </div>

            <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title="Новый агрегат">
                <div className="form-row">
                    <label className="form-label">Название агрегата</label>
                    <input type="text" className="form-input" value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })} />
                </div>
                <div className="form-row">
                    <label className="form-label">Описание</label>
                    <textarea className="form-input" rows="2" value={formData.description} onChange={(e) => setFormData({ ...formData, description: e.target.value })} />
                </div>
                <div className="form-row">
                    <label className="form-label">Числовые параметры</label>
                    <div style={{ maxHeight: '150px', overflow: 'auto', border: '1px solid var(--border)', borderRadius: '8px', padding: '8px' }}>
                        {params.numeric.map(p => (
                            <label key={p.id} className="form-check" style={{ display: 'block', marginBottom: '4px' }}>
                                <input type="checkbox" checked={formData.numeric_param_ids.includes(p.id)}
                                       onChange={(e) => {
                                           if (e.target.checked) setFormData({ ...formData, numeric_param_ids: [...formData.numeric_param_ids, p.id] })
                                           else setFormData({ ...formData, numeric_param_ids: formData.numeric_param_ids.filter(id => id !== p.id) })
                                       }} />
                                {p.field_name} ({p.min_val}—{p.max_val})
                            </label>
                        ))}
                    </div>
                </div>
                <div className="form-row">
                    <label className="form-label">Enum параметры</label>
                    <div style={{ maxHeight: '150px', overflow: 'auto', border: '1px solid var(--border)', borderRadius: '8px', padding: '8px' }}>
                        {params.enum.map(p => (
                            <label key={p.id} className="form-check" style={{ display: 'block', marginBottom: '4px' }}>
                                <input type="checkbox" checked={formData.enum_param_ids.includes(p.id)}
                                       onChange={(e) => {
                                           if (e.target.checked) setFormData({ ...formData, enum_param_ids: [...formData.enum_param_ids, p.id] })
                                           else setFormData({ ...formData, enum_param_ids: formData.enum_param_ids.filter(id => id !== p.id) })
                                       }} />
                                Enum #{p.enum_class_id}
                            </label>
                        ))}
                    </div>
                </div>
                <div className="modal-footer">
                    <button className="btn btn-ghost" onClick={() => setModalOpen(false)}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleCreate}>Создать</button>
                </div>
            </Modal>
        </div>
    )
}