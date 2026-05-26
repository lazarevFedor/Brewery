import { useState, useEffect } from 'react'
import { beersApi, categoriesApi } from '../api/client'
import { useToast } from '../hooks/useToast'
import { Modal } from '../components/Modal'
import '../styles/admin.css'

const emptyForm = {
    name: '',
    description: '',
    abv: '',
    ibu: '',
    amount: '',
    units: 'мл',
    city: '',
    country: '',
    categoryName: '',
    features: '',
}

export function AdminBeers() {
    const [beers, setBeers] = useState([])
    const [categories, setCategories] = useState([])
    const [modalOpen, setModalOpen] = useState(false)
    const [editingBeer, setEditingBeer] = useState(null)
    const [form, setForm] = useState(emptyForm)
    const [search, setSearch] = useState('')
    const { showToast } = useToast()

    useEffect(() => {
        loadBeers()
        loadCategories()
    }, [])

    const loadBeers = async () => {
        try {
            const data = await beersApi.getAll(0, 500)
            setBeers(data || [])
        } catch {
            showToast('Ошибка загрузки пива', 'error')
        }
    }

    const loadCategories = async () => {
        try {
            const data = await categoriesApi.getAll()
            setCategories(data || [])
        } catch {}
    }

    const openCreate = () => {
        setEditingBeer(null)
        setForm(emptyForm)
        setModalOpen(true)
    }

    const openEdit = (beer) => {
        setEditingBeer(beer)
        setForm({
            name: beer.name || '',
            description: beer.description || '',
            abv: beer.abv ?? '',
            ibu: beer.ibu ?? '',
            amount: beer.amount ?? '',
            units: beer.units || 'мл',
            city: beer.city || '',
            country: beer.country || '',
            categoryName: beer.category?.name || '',
            features: (beer.features || []).join('\n'),
        })
        setModalOpen(true)
    }

    const handleSave = async () => {
        if (!form.name.trim()) {
            showToast('Введите название пива', 'error')
            return
        }

        // Base fields that map directly to beers table columns.
        // features are in beer_features join table (not a beers column — sending them in PATCH causes 500).
        // city/country are resolved server-side to city_id; omit if empty to avoid "city name is empty" error.
        const payload = {
            name: form.name,
            description: form.description,
            abv: form.abv !== '' ? parseFloat(form.abv) : 0,
            ibu: form.ibu !== '' ? parseFloat(form.ibu) : 0,
            amount: form.amount !== '' ? parseInt(form.amount) : 0,
            units: form.units,
        }

        if (editingBeer) {
            // PATCH: only include optional fields when present
            if (form.categoryName) payload.category = { name: form.categoryName }
            if (form.city && form.country) {
                payload.city = form.city
                payload.country = form.country
            }
        } else {
            // POST: city, country, category are required by InsertBeer
            if (!form.city || !form.country) {
                showToast('Укажите город и страну', 'error')
                return
            }
            if (!form.categoryName) {
                showToast('Выберите категорию', 'error')
                return
            }
            payload.city = form.city
            payload.country = form.country
            payload.category = { name: form.categoryName }
            payload.features = form.features ? form.features.split('\n').map(s => s.trim()).filter(Boolean) : []
        }

        try {
            if (editingBeer) {
                await beersApi.update(editingBeer.id, payload)
                showToast('Пиво обновлено', 'success')
            } else {
                await beersApi.create(payload)
                showToast('Пиво добавлено', 'success')
            }
            setModalOpen(false)
            loadBeers()
        } catch (e) {
            showToast(e.message || 'Ошибка сохранения', 'error')
        }
    }

    const handleDelete = async (id, e) => {
        e.stopPropagation()
        if (!confirm('Удалить пиво?')) return
        try {
            await beersApi.delete(id)
            showToast('Пиво удалено', 'success')
            loadBeers()
        } catch {
            showToast('Ошибка удаления', 'error')
        }
    }

    const visible = beers.filter(b =>
        b.name?.toLowerCase().includes(search.toLowerCase()) ||
        b.country?.toLowerCase().includes(search.toLowerCase())
    )

    return (
        <div>
            <div className="admin-header">
                <div>
                    <div className="admin-title">Пиво</div>
                    <div className="admin-sub">{beers.length} позиций в каталоге</div>
                </div>
                <button className="btn btn-primary" onClick={openCreate}>+ Добавить пиво</button>
            </div>

            <div style={{ marginBottom: '12px' }}>
                <input
                    type="text"
                    className="form-input"
                    placeholder="Поиск по названию или стране..."
                    value={search}
                    onChange={e => setSearch(e.target.value)}
                    style={{ maxWidth: '320px' }}
                />
            </div>

            <div className="table-wrap">
                <table>
                    <thead>
                        <tr>
                            <th>Название</th>
                            <th>Страна</th>
                            <th style={{ width: 70, textAlign: 'center' }}>ABV%</th>
                            <th style={{ width: 70, textAlign: 'center' }}>IBU</th>
                            <th style={{ width: 90, textAlign: 'center' }}>Количество</th>
                            <th>Категория</th>
                            <th style={{ width: 80 }}></th>
                        </tr>
                    </thead>
                    <tbody>
                        {visible.length === 0 && (
                            <tr><td colSpan={7} style={{ textAlign: 'center', color: 'var(--text-hint)', padding: '24px' }}>
                                Нет позиций
                            </td></tr>
                        )}
                        {visible.map(beer => (
                            <tr key={beer.id} style={{ cursor: 'pointer' }} onClick={() => openEdit(beer)}>
                                <td style={{ fontWeight: 500 }}>{beer.name}</td>
                                <td>{beer.country || <span style={{ color: 'var(--text-hint)' }}>—</span>}</td>
                                <td style={{ textAlign: 'center' }}>{beer.abv ?? '—'}</td>
                                <td style={{ textAlign: 'center' }}>{beer.ibu ?? '—'}</td>
                                <td style={{ textAlign: 'center' }}>
                                    {beer.amount
                                        ? <>{beer.amount} {beer.units}</>
                                        : <span style={{ color: 'var(--text-hint)' }}>—</span>}
                                </td>
                                <td>{beer.category?.name || <span style={{ color: 'var(--text-hint)' }}>—</span>}</td>
                                <td>
                                    <button className="icon-btn del" onClick={(e) => handleDelete(beer.id, e)}>🗑</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title={editingBeer ? 'Редактировать пиво' : 'Новое пиво'}>
                <div className="form-row">
                    <label className="form-label">Название *</label>
                    <input type="text" className="form-input" autoFocus value={form.name}
                        onChange={e => setForm({ ...form, name: e.target.value })} />
                </div>
                <div className="form-row">
                    <label className="form-label">Описание</label>
                    <textarea className="form-input" rows="3" value={form.description}
                        onChange={e => setForm({ ...form, description: e.target.value })} />
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                    <div className="form-row">
                        <label className="form-label">ABV (%)</label>
                        <input type="number" step="0.1" className="form-input" value={form.abv}
                            onChange={e => setForm({ ...form, abv: e.target.value })} />
                    </div>
                    <div className="form-row">
                        <label className="form-label">IBU</label>
                        <input type="number" className="form-input" value={form.ibu}
                            onChange={e => setForm({ ...form, ibu: e.target.value })} />
                    </div>
                    <div className="form-row">
                        <label className="form-label">Количество</label>
                        <input type="number" className="form-input" value={form.amount}
                            onChange={e => setForm({ ...form, amount: e.target.value })} />
                    </div>
                    <div className="form-row">
                        <label className="form-label">Единицы</label>
                        <input type="text" className="form-input" value={form.units}
                            onChange={e => setForm({ ...form, units: e.target.value })} placeholder="мл, л, г..." />
                    </div>
                    <div className="form-row">
                        <label className="form-label">Страна{!editingBeer && ' *'}</label>
                        <input type="text" className="form-input" value={form.country}
                            onChange={e => setForm({ ...form, country: e.target.value })} />
                    </div>
                    <div className="form-row">
                        <label className="form-label">Город{!editingBeer && ' *'}</label>
                        <input type="text" className="form-input" value={form.city}
                            onChange={e => setForm({ ...form, city: e.target.value })} />
                    </div>
                </div>
                <div className="form-row">
                    <label className="form-label">Категория{!editingBeer && ' *'}</label>
                    <select className="form-select" value={form.categoryName}
                        onChange={e => setForm({ ...form, categoryName: e.target.value })}>
                        <option value="">— без категории —</option>
                        {categories.map(cat => (
                            <option key={cat.id} value={cat.name}>{cat.name}</option>
                        ))}
                    </select>
                </div>
                <div className="form-row">
                    <label className="form-label">Особенности (каждая с новой строки)</label>
                    <textarea className="form-input" rows="3" value={form.features}
                        onChange={e => setForm({ ...form, features: e.target.value })}
                        placeholder="Нефильтрованное&#10;Непастеризованное&#10;Живое" />
                </div>
                <div className="modal-footer">
                    <button className="btn btn-ghost" onClick={() => setModalOpen(false)}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleSave}>
                        {editingBeer ? 'Сохранить' : 'Создать'}
                    </button>
                </div>
            </Modal>
        </div>
    )
}
