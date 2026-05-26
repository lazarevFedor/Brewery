import { useState, useEffect } from 'react'
import { parametersApi, enumApi } from '../api/client'
import { useToast } from '../hooks/useToast'
import { Modal } from '../components/Modal'
import './styles/admin.css'

export function AdminParams() {
    const [params, setParams] = useState([])
    const [enumClasses, setEnumClasses] = useState([])
    const [modalOpen, setModalOpen] = useState(false)
    const [paramType, setParamType] = useState('numeric')
    const [formData, setFormData] = useState({
        field_name: '',
        entity_name: 'beers',
        min_val: 0,
        max_val: 100,
        enum_class_id: 0,
        inheritable: true,
    })
    const { showToast } = useToast()

    useEffect(() => {
        loadParams()
        loadEnumClasses()
    }, [])

    const loadParams = async () => {
        try {
            const data = await parametersApi.getAll()
            setParams(data)
        } catch {
            showToast('Ошибка загрузки параметров', 'error')
        }
    }

    const loadEnumClasses = async () => {
        try {
            const data = await enumApi.getAll()
            setEnumClasses(data)
        } catch {
            // ignore
        }
    }

    const handleCreate = async () => {
        try {
            if (paramType === 'numeric') {
                await parametersApi.createNumeric({
                    field_name: formData.field_name,
                    entity_name: formData.entity_name,
                    min_val: formData.min_val,
                    max_val: formData.max_val,
                    inheritable: formData.inheritable,
                })
            } else {
                if (!formData.enum_class_id) {
                    showToast('Выберите Enum-класс', 'error')
                    return
                }
                await parametersApi.createEnum({
                    enum_class_id: formData.enum_class_id,
                    inheritable: formData.inheritable,
                })
            }
            showToast('Параметр создан', 'success')
            setModalOpen(false)
            loadParams()
            setFormData({
                field_name: '',
                entity_name: 'beers',
                min_val: 0,
                max_val: 100,
                enum_class_id: 0,
                inheritable: true,
            })
        } catch {
            showToast('Ошибка создания', 'error')
        }
    }

    const handleDelete = async (param) => {
        if (!confirm('Удалить параметр?')) return
        try {
            if ('min_val' in param) {
                await parametersApi.deleteNumeric(param.id)
            } else {
                await parametersApi.deleteEnum(param.id)
            }
            showToast('Параметр удалён', 'success')
            loadParams()
        } catch {
            showToast('Ошибка удаления', 'error')
        }
    }

    const renderRange = (param) => {
        if ('min_val' in param) {
            return `${param.min_val} — ${param.max_val}`
        }
        const enumClass = enumClasses.find((c) => c.id === param.enum_class_id)
        return enumClass?.enum_type || `Enum #${param.enum_class_id}`
    }

    return (
        <div>
            <div className="admin-header">
                <div>
                    <div className="admin-title">Параметры категорий</div>
                    <div className="admin-sub">Управление параметрами для категорий пива</div>
                </div>
                <button className="btn btn-primary" onClick={() => setModalOpen(true)}>
                    + Добавить параметр
                </button>
            </div>

            <div className="table-wrap">
                <table>
                    <thead>
                    <tr><th>Поле</th><th>Тип</th><th>Диапазон / Enum-класс</th><th>Наследуется</th><th></th></tr>
                    </thead>
                    <tbody>
                    {params.map((param) => (
                        <tr key={param.id}>
                            <td><strong>{param.field_name || `Enum #${param.id}`}</strong></td>
                            <td><span className={`badge ${'min_val' in param ? 'badge-numeric' : 'badge-enum'}`}>
                  {'min_val' in param ? 'numeric' : 'enum'}
                </span></td>
                            <td style={{ color: '#666' }}>{renderRange(param)}</td>
                            <td><span className={`badge ${param.inheritable ? 'badge-yes' : 'badge-no'}`}>
                  {param.inheritable ? 'да' : 'нет'}
                </span></td>
                            <td><button className="icon-btn del" onClick={() => handleDelete(param)}>🗑</button></td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>

            <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title="Новый параметр">
                <div className="form-row">
                    <label className="form-label">Поле (field name)</label>
                    <input type="text" className="form-input" placeholder="например: abv, ibu, country"
                           value={formData.field_name} onChange={(e) => setFormData({ ...formData, field_name: e.target.value })} />
                </div>
                <div className="form-row">
                    <label className="form-label">Тип</label>
                    <select className="form-select" value={paramType} onChange={(e) => setParamType(e.target.value)}>
                        <option value="numeric">numeric</option>
                        <option value="enum">enum</option>
                    </select>
                </div>

                {paramType === 'numeric' && (
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
                        <input type="number" className="form-input" placeholder="Минимум" value={formData.min_val}
                               onChange={(e) => setFormData({ ...formData, min_val: parseInt(e.target.value) || 0 })} />
                        <input type="number" className="form-input" placeholder="Максимум" value={formData.max_val}
                               onChange={(e) => setFormData({ ...formData, max_val: parseInt(e.target.value) || 0 })} />
                    </div>
                )}

                {paramType === 'enum' && (
                    <select className="form-select" value={formData.enum_class_id}
                            onChange={(e) => setFormData({ ...formData, enum_class_id: parseInt(e.target.value) })}>
                        <option value={0}>Выберите Enum-класс</option>
                        {enumClasses.map((cls) => (
                            <option key={cls.id} value={cls.id}>{cls.enum_type}</option>
                        ))}
                    </select>
                )}

                <div className="form-row">
                    <label className="form-check">
                        <input type="checkbox" checked={formData.inheritable}
                               onChange={(e) => setFormData({ ...formData, inheritable: e.target.checked })} />
                        Наследуется дочерними категориями
                    </label>
                </div>

                <div className="modal-footer">
                    <button className="btn btn-ghost" onClick={() => setModalOpen(false)}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleCreate}>Сохранить</button>
                </div>
            </Modal>
        </div>
    )
}