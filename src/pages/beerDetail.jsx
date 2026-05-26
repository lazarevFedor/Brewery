import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { beersApi, reviewsApi } from '../api/client'
import { useToast } from '../hooks/useToast'
import './styles/detail.css'

export function BeerDetail() {
    const { id } = useParams()
    const navigate = useNavigate()
    const [beer, setBeer] = useState(null)
    const [reviews, setReviews] = useState([])
    const [reviewStars, setReviewStars] = useState(0)
    const [reviewName, setReviewName] = useState('')
    const [reviewText, setReviewText] = useState('')
    const { showToast } = useToast()

    useEffect(() => {
        loadBeer()
        loadReviews()
    }, [id])

    const loadBeer = async () => {
        try {
            const data = await beersApi.getById(id)
            setBeer(data)
        } catch {
            navigate('/catalog')
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
        if (!reviewName || !reviewText || reviewStars === 0) {
            showToast('Заполните все поля и выберите оценку', 'error')
            return
        }

        try {
            await reviewsApi.create(id, {
                author: reviewName,
                stars: reviewStars,
                text: reviewText
            })
            showToast('Спасибо! Отзыв добавлен.', 'success')
            setReviewName('')
            setReviewText('')
            setReviewStars(0)
            loadReviews()
        } catch {
            showToast('Ошибка при добавлении отзыва', 'error')
        }
    }

    if (!beer) return <div>Загрузка...</div>

    return (
        <div>
            <div className="breadcrumb">
                <a href="#" onClick={() => navigate('/catalog')}>Все</a>
                <span className="sep">›</span>
                <span style={{ fontWeight: 600 }}>{beer.name}</span>
            </div>

            <div className="detail-grid">
                <div className="beer-hero">🍺</div>
                <div className="detail-info">
                    <div className="detail-title">{beer.name}</div>
                    <div className="detail-cat">{beer.style} · {beer.country}</div>

                    <div className="feat-grid">
                        <div className="feat-card">
                            <div className="feat-card-label">Крепость</div>
                            <div className="feat-card-val">{beer.abv}<span className="feat-card-unit">%</span></div>
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
                            <div className="feat-card-label">Стиль</div>
                            <div className="feat-card-val" style={{ fontSize: '14px', marginTop: '6px' }}>{beer.style}</div>
                        </div>
                    </div>

                    <div className="btn-row">
                        <button className="btn btn-primary" onClick={() => document.getElementById('review-form').scrollIntoView({ behavior: 'smooth' })}>
                            ✍️ Оставить отзыв
                        </button>
                    </div>

                    <div className="section-title">Отзывы</div>
                    <div className="reviews-list">
                        {reviews.length === 0 && <p className="text-hint">Отзывов пока нет. Будьте первым!</p>}
                        {reviews.map((review, idx) => (
                            <div key={idx} className="review-item">
                                <div className="review-head">
                                    <div className="review-author">{review.author}</div>
                                    <div className="stars">{'★'.repeat(review.stars)}{'☆'.repeat(5 - review.stars)}</div>
                                </div>
                                <div className="review-text">{review.text}</div>
                            </div>
                        ))}
                    </div>

                    <div className="review-form-box" id="review-form">
                        <div className="section-title" style={{ border: 'none', paddingBottom: 0 }}>Ваш отзыв</div>
                        <label className="form-label">Оценка</label>
                        <div className="star-pick">
                            {[1, 2, 3, 4, 5].map(star => (
                                <span key={star} onClick={() => setReviewStars(star)} className={star <= reviewStars ? 'lit' : ''}>★</span>
                            ))}
                        </div>
                        <label className="form-label">Имя</label>
                        <input type="text" className="form-input" value={reviewName} onChange={(e) => setReviewName(e.target.value)} placeholder="Ваше имя..." />
                        <label className="form-label">Комментарий</label>
                        <textarea className="form-input" rows="3" value={reviewText} onChange={(e) => setReviewText(e.target.value)} placeholder="Что думаете об этом пиве?"></textarea>
                        <button className="btn btn-primary" onClick={submitReview}>Отправить отзыв</button>
                    </div>
                </div>
            </div>
        </div>
    )
}