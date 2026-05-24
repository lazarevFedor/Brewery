import './styles/promotions.css'

const promotions = [
    { id: 1, icon: '🎄', discount: '-25%', badge: 'Сезонная', title: 'Новогодний сет', desc: 'При покупке от 3 литров — скидка 25% на весь заказ.', date: 'до 15 января 2026', code: 'НОВЫЙГОД25' },
    { id: 2, icon: '👥', discount: '-15%', badge: 'Друзьям', title: 'Приведи друга', desc: 'Порекомендуй нас другу — получи скидку 15% на следующий заказ.', date: 'бессрочно', code: 'BEERFRIEND' },
    { id: 3, icon: '📦', discount: 'Бесплатно', badge: 'Доставка', title: 'Бесплатная доставка', desc: 'При заказе от 3000 рублей — доставка бесплатно.', date: 'до 31 декабря 2026', code: 'Автоматически в корзине' },
    { id: 4, icon: '🎂', discount: '-20%', badge: 'Именинникам', title: 'Скидка в день рождения', desc: 'В день вашего рождения — персональная скидка 20%.', date: 'в день рождения', code: 'HAPPYBIRTHDAY' },
    { id: 5, icon: '🍻', discount: '6+1', badge: 'Наборы', title: 'Купи 6 — получи 7-е в подарок', desc: 'При покупке 6 бутылок, седьмая — в подарок.', date: 'до конца месяца', code: 'Автоматически в корзине' },
    { id: 6, icon: '⭐', discount: '-10%', badge: 'Новинки', title: 'Скидка на новинки', desc: 'Попробуйте новые сорта со скидкой 10%.', date: 'на новинки', code: 'NEWBEER10' },
]

export function Promotions() {
    return (
        <>
            <div className="promo-header">
                <h1>🍺 Акции и скидки</h1>
                <p>Пейте с выгодой — только лучшие предложения для вас!</p>
            </div>

            <div className="promo-grid">
                {promotions.map(promo => (
                    <div key={promo.id} className="promo-card">
                        <div className="promo-image">{promo.icon}</div>
                        <div className="promo-content">
                            <span className="discount-badge">{promo.discount}</span>
                            <span className="promo-badge">{promo.badge}</span>
                            <div className="promo-title">{promo.title}</div>
                            <div className="promo-desc">{promo.desc}</div>
                            <div className="promo-date">📅 {promo.date}</div>
                            <div className="promo-code">{promo.code}</div>
                        </div>
                    </div>
                ))}
            </div>

            <div className="info-section">
                <h3>🎁 Бонусная программа</h3>
                <p>Каждая покупка приносит вам бонусные баллы:</p>
                <ul style={{ marginTop: '12px', marginLeft: '20px' }}>
                    <li>1 балл = 1 рубль скидки</li>
                    <li>5% от суммы заказа начисляется на бонусный счёт</li>
                    <li>100 баллов за первый заказ</li>
                </ul>
            </div>

            <div className="bonus-card">
                <div style={{ fontSize: '32px' }}>🏆</div>
                <div style={{ fontFamily: "'Unbounded', sans-serif", fontWeight: 700, marginBottom: '12px' }}>VIP-статус</div>
                <div>Накопите 5000 баллов и получите пожизненную скидку 15% + именную кружку!</div>
            </div>

            <div className="info-section" style={{ marginTop: '20px' }}>
                <h3>📜 Условия акций</h3>
                <p style={{ fontSize: '12px', color: 'var(--text-hint)' }}>* Акции не суммируются. Применяется максимальная скидка. Промокоды одноразовые. Бонусы действуют 90 дней.</p>
            </div>
        </>
    )
}