import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { beersApi, reviewsApi } from '../api/client'
import '../styles/detail.css'

export function BeerDetail() {
    const { id } = useParams()
    const navigate = useNavigate()
    const [beer, setBeer] = useState(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [reviews, setReviews] = useState([])
    const [reviewStars, setReviewStars] = useState(0)
    const [reviewName, setReviewName] = useState('')
    const [reviewText, setReviewText] = useState('')
    const [reviewAnon, setReviewAnon] = useState(false)
    const [toast, setToast] = useState({ msg: '', type: '' })

    const showToast = (msg, type = 'success') => {
        setToast({ msg, type })
        setTimeout(() => setToast({ msg: '', type: '' }), 3000)
    }

    useEffect(() => {
        loadBeer()
        loadReviews()
    }, [id])

    const loadBeer = async () => {
        try {
            const data = await beersApi.getById(id)
            setBeer(data)
        } catch (err) {
            console.error('Error loading beer:', err)
            setError(err.message || 'Не удалось загрузить пиво')
        } finally {
            setLoading(false)
        }
    }

    const loadReviews = async () => {
        try {
            const data = await reviewsApi.getByBeerId(id)
            setReviews(data)
        } catch {
            // ignore
        }
    }

    const submitReview = async () => {
        if ((!reviewAnon && !reviewName) || !reviewText || reviewStars === 0) {
            showToast('Заполните все поля и выберите оценку', 'error')
            return
        }
        try {
            await reviewsApi.create(id, {
                author: reviewAnon ? '' : reviewName,
                rating: reviewStars,
                body: reviewText
            })
            setReviewName('')
            setReviewText('')
            setReviewStars(0)
            setReviewAnon(false)
            loadReviews()
            showToast('Спасибо за отзыв!')
        } catch (error) {
            console.error('Error submitting review:', error)
            showToast('Ошибка при добавлении отзыва', 'error')
        }
    }

    const setStars = (n) => {
        setReviewStars(n)
        const spans = document.querySelectorAll('#star-pick span')
        spans.forEach((s, i) => {
            if (i < n) {
                s.classList.add('lit')
            } else {
                s.classList.remove('lit')
            }
        })
    }

    if (loading) {
        return <div className="text-hint" style={{ textAlign: 'center', padding: '40px' }}>Загрузка...</div>
    }

    if (error) {
        return (
            <div style={{ textAlign: 'center', padding: '40px' }}>
                <p style={{ color: '#c0392b', marginBottom: '16px' }}>Ошибка: {error}</p>
                <button onClick={() => navigate('/catalog')} className="btn btn-primary">← Вернуться в каталог</button>
            </div>
        )
    }

    if (!beer) {
        return null
    }

    return (
        <div>
            <div className="breadcrumb">
                <a onClick={() => navigate('/catalog')} style={{ cursor: 'pointer' }}>Все</a>
                <span className="sep">›</span>
                <a onClick={() => navigate('/catalog')} style={{ cursor: 'pointer' }}>{beer.category?.name || 'Пиво'}</a>
                <span className="sep">›</span>
                <span style={{ fontWeight: 600 }}>{beer.name}</span>
            </div>

            <div className="detail-grid">
                <div className="beer-hero">🍺</div>
                <div className="detail-info">
                    <div className="detail-title">{beer.name}</div>
                    <div className="detail-cat">{beer.category?.name || 'Пиво'} › {beer.city ? `${beer.city}, ` : ''}{beer.country}</div>

                    {beer.description && (
                        <p style={{ fontSize: '13px', color: 'var(--text-muted)', lineHeight: 1.6, marginBottom: '16px' }}>
                            {beer.description}
                        </p>
                    )}

                    <div className="feat-grid">
                        <div className="feat-card">
                            <div className="feat-card-label">Крепость</div>
                            <div className="feat-card-val">{beer.abv}<span className="feat-card-unit">% ABV</span></div>
                        </div>
                        <div className="feat-card">
                            <div className="feat-card-label">Горечь</div>
                            <div className="feat-card-val">{beer.ibu}<span className="feat-card-unit"> IBU</span></div>
                        </div>
                        <div className="feat-card">
                            <div className="feat-card-label">Страна</div>
                            <div className="feat-card-val" style={{ fontSize: '14px', marginTop: '6px' }}>{beer.country}</div>
                        </div>
                        <div className="feat-card">
                            <div className="feat-card-label">В наличии</div>
                            <div className="feat-card-val" style={{ fontSize: '14px', marginTop: '6px' }}>
                                {beer.amount} <span style={{ fontSize: '11px', fontWeight: 400 }}>{beer.units}</span>
                            </div>
                        </div>
                    </div>

                    {beer.features?.length > 0 && (
                        <div style={{ marginBottom: '20px' }}>
                            <div style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text-hint)', textTransform: 'uppercase', marginBottom: '8px' }}>Особенности</div>
                            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px' }}>
                                {beer.features.map((f, i) => (
                                    <span key={i} className="badge badge-enum">{f}</span>
                                ))}
                            </div>
                        </div>
                    )}

                    <div className="btn-row">
                        <button
                            className="btn btn-primary"
                            onClick={() => document.getElementById('review-form')?.scrollIntoView({ behavior: 'smooth' })}
                        >
                            ✍️ Оставить отзыв
                        </button>
                    </div>

                    <div className="section-title" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <span>Отзывы</span>
                        {reviews.filter(r => r.rating > 0).length > 0 && (() => {
                            const rated = reviews.filter(r => r.rating > 0)
                            return (
                                <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--amber)' }}>
                                    ★ {(rated.reduce((s, r) => s + r.rating, 0) / rated.length).toFixed(1)}
                                    <span style={{ fontWeight: 400, color: 'var(--text-hint)', marginLeft: '4px' }}>
                                        ({reviews.length})
                                    </span>
                                </span>
                            )
                        })()}
                    </div>
                    <div className="reviews-list">
                        {reviews.length === 0 && (
                            <p className="text-hint" style={{ padding: '12px 0' }}>Отзывов пока нет. Будьте первым!</p>
                        )}
                        {reviews.map((review, idx) => (
                            <div key={idx} className="review-item">
                                <div className="review-head">
                                    <div className="review-author">{review.author || 'Аноним'}</div>
                                    <div className="stars">
                                        {'★'.repeat(review.rating)}
                                        {'☆'.repeat(5 - review.rating)}
                                    </div>
                                </div>
                                <div className="review-text">{review.body || review.text}</div>
                            </div>
                        ))}
                    </div>

                    <div className="review-form-box" id="review-form">
                        <div className="section-title" style={{ border: 'none', paddingBottom: 0 }}>Ваш отзыв</div>
                        <label className="form-label">Оценка</label>
                        <div className="star-pick" id="star-pick">
                            <span onClick={() => setStars(1)}>★</span>
                            <span onClick={() => setStars(2)}>★</span>
                            <span onClick={() => setStars(3)}>★</span>
                            <span onClick={() => setStars(4)}>★</span>
                            <span onClick={() => setStars(5)}>★</span>
                        </div>
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '6px' }}>
                            <label className="form-label" style={{ margin: 0 }}>Имя</label>
                            <label className="form-check" style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                                <input
                                    type="checkbox"
                                    checked={reviewAnon}
                                    onChange={(e) => setReviewAnon(e.target.checked)}
                                />
                                Анонимно
                            </label>
                        </div>
                        {!reviewAnon && (
                            <input
                                type="text"
                                className="form-input"
                                placeholder="Ваше имя..."
                                value={reviewName}
                                onChange={(e) => setReviewName(e.target.value)}
                            />
                        )}
                        <label className="form-label">Комментарий</label>
                        <textarea
                            className="form-input"
                            placeholder="Что думаете об этом пиве?"
                            rows="3"
                            value={reviewText}
                            onChange={(e) => setReviewText(e.target.value)}
                        />
                        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                            <button className="btn btn-primary" onClick={submitReview}>Отправить</button>
                        </div>
                    </div>
                </div>
            </div>
            <div className={`toast ${toast.type} ${toast.msg ? 'show' : ''}`}>{toast.msg}</div>
        </div>
    )
}