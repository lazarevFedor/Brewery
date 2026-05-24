import { useState, useEffect } from 'react'
import { enumApi, enumValueApi } from '../api/client'
import { useToast } from '../hooks/useToast'
import { Modal } from '../components/Modal'
import './styles/admin.css'

export function AdminEnums() {
    const [classes, setClasses] = useState([])
    const [selectedClass, setSelectedClass] = useState(null)
    const [values, setValues] = useState([])
    const [modalOpen, setModalOpen] = useState(false)
    const [newClassName, setNewClassName] = useState('')
    const [newValue, setNewValue] = useState('')
    const { showToast } = useToast()

    useEffect(() => {
        loadClasses()
    }, [])

    useEffect(() => {
        if (selectedClass) loadValues()
    }, [selectedClass])

    const loadClasses = async () => {
        try {
            const data = await enumApi.getAll()
            setClasses(data)
            if (data.length > 0 && !selectedClass) setSelectedClass(data[0])
        } catch {
            showToast('Ошибка загрузки классов', 'error')
        }
    }

    const loadValues = async () => {
        if (!selectedClass) return
        try {
            const data = await enumValueApi.getAll(selectedClass.entity_name, selectedClass.field_name, selectedClass.enum_type)
            setValues(data)
        } catch {
            showToast('Ошибка загрузки значений', 'error')
        }
    }

    const handleAddClass = async () => {
        if (!newClassName.trim()) return
        try {
            await enumApi.create({
                enum_type: newClassName,
                entity_name: 'beers',
                field_name: newClassName.toLowerCase(),
                is_active: true,
            })
            showToast('Класс создан', 'success')
            setModalOpen(false)
            setNewClassName('')
            loadClasses()
        } catch {
            showToast('Ошибка создания', 'error')
        }
    }

    const handleAddValue = async () => {
        if (!selectedClass || !newValue.trim()) return
        try {
            await enumValueApi.create({
                class_id: selectedClass.id,
                enum_type: selectedClass.enum_type,
                value: newValue,
                position: values.length + 1,
            })
            showToast('Значение добавлено', 'success')
            setNewValue('')
            loadValues()
        } catch {
            showToast('Ошибка добавления', 'error')
        }
    }

    const handleDeleteValue = async (id) => {
        if (!confirm('Удалить значение?')) return
        try {
            await enumValueApi.delete(id)
            showToast('Значение удалено', 'success')
            loadValues()
        } catch {
            showToast('Ошибка удаления', 'error')
        }
    }

    const handleDeleteClass = async (id) => {
        if (!confirm('Удалить класс? Все значения также будут удалены.')) return
        try {
            await enumApi.delete(id)
            showToast('Класс удалён', 'success')
            if (selectedClass?.id === id) setSelectedClass(null)
            loadClasses()
        } catch {
            showToast('Ошибка удаления', 'error')
        }
    }

    return (
        <div>
            <div className="admin-header">
                <div>
                    <div className="admin-title">Enum-справочники</div>
                    <div className="admin-sub">Управление перечислениями значений</div>
                </div>
                <button className="btn btn-primary" onClick={() => setModalOpen(true)}>+ Новый класс</button>
            </div>

            <div className="enum-layout">
                <div className="enum-list">
                    {classes.map((cls) => (
                        <div key={cls.id} className={`enum-class-item ${selectedClass?.id === cls.id ? 'active' : ''}`}
                             onClick={() => setSelectedClass(cls)}>
                            <span>{cls.enum_type}</span>
                            <div>
                                <span className="cnt">{values.length} зн.</span>
                                <button className="icon-btn" onClick={(e) => { e.stopPropagation(); handleDeleteClass(cls.id) }} style={{ marginLeft: '8px' }}>🗑</button>
                            </div>
                        </div>
                    ))}
                </div>

                <div>
                    <div className="admin-header" style={{ marginBottom: '12px' }}>
                        <div>
                            <div style={{ fontSize: '14px', fontWeight: 600 }}>{selectedClass?.enum_type || 'Выберите класс'}</div>
                            <div className="admin-sub">{values.length} значений</div>
                        </div>
                    </div>
                    <div className="table-wrap">
                        <table>
                            <thead><tr><th style={{ width: '60px' }}>ID</th><th>Значение</th><th style={{ width: '70px' }}></th></tr></thead>
                            <tbody>
                            {values.map((val, idx) => (
                                <tr key={val.id}><td style={{ color: '#999' }}>{idx + 1}</td><td>{val.value}</td>
                                    <td><button className="icon-btn del" onClick={() => handleDeleteValue(val.id)}>🗑</button></td></tr>
                            ))}
                            </tbody>
                        </table>
                        <div className="add-val-row">
                            <input type="text" placeholder="Новое значение..." value={newValue} onChange={(e) => setNewValue(e.target.value)}
                                   onKeyDown={(e) => e.key === 'Enter' && handleAddValue()} />
                            <button className="btn btn-primary" onClick={handleAddValue}>Добавить</button>
                        </div>
                    </div>
                </div>
            </div>

            <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title="Новый Enum-класс">
                <div className="form-row">
                    <label className="form-label">Название класса</label>
                    <input type="text" className="form-input" value={newClassName} onChange={(e) => setNewClassName(e.target.value)} placeholder="например: CountryEnum" />
                </div>
                <div className="modal-footer">
                    <button className="btn btn-ghost" onClick={() => setModalOpen(false)}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleAddClass}>Создать</button>
                </div>
            </Modal>
        </div>
    )
}