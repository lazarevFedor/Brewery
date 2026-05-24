import './styles/contacts.css'

export function Contacts() {
    return (
        <>
            <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
                <h1 style={{ fontSize: '32px', marginBottom: '0.5rem' }}>📍 Контакты</h1>
                <p style={{ color: 'var(--text-muted)', fontSize: '16px' }}>Всегда рады помочь и ответить на ваши вопросы</p>
            </div>

            <div className="contacts-grid">
                <div className="contact-card">
                    <div className="contact-icon">📞</div>
                    <h3 style={{ fontSize: '16px', marginBottom: '8px' }}>Телефон</h3>
                    <p style={{ fontSize: '20px', fontWeight: 700, color: 'var(--accent)' }}>+7 (495) 123-45-67</p>
                    <p className="text-hint">Пн-Вс: 10:00–21:00</p>
                </div>
                <div className="contact-card">
                    <div className="contact-icon">✉️</div>
                    <h3 style={{ fontSize: '16px', marginBottom: '8px' }}>Email</h3>
                    <p><a href="mailto:info@beer-catalog.ru" style={{ color: 'var(--blue)', textDecoration: 'none' }}>info@beer-catalog.ru</a></p>
                    <p><a href="mailto:order@beer-catalog.ru" style={{ color: 'var(--blue)', textDecoration: 'none' }}>order@beer-catalog.ru</a></p>
                </div>
                <div className="contact-card">
                    <div className="contact-icon">📍</div>
                    <h3 style={{ fontSize: '16px', marginBottom: '8px' }}>Адрес</h3>
                    <p>Москва, ул. Пивоваров, д. 15</p>
                    <p className="text-hint">м. Курская, выход №4</p>
                </div>
            </div>

            <div className="order-phone">
                🚚 <strong>Бесплатная доставка</strong> от 3000₽ — звоните <a href="tel:+74951234567">+7 (495) 123-45-67</a>
            </div>

            <div className="map-placeholder">
                🗺️ Карта проезда (интерактивная карта будет здесь)
            </div>
        </>
    )
}