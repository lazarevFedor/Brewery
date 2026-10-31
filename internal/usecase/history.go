package usecase

import (
	"Brewery/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HistoryRecordRepository определяет интерфейс для работы с записями истории в базе данных.
type HistoryRecordRepository interface {
}

// NewHistoryRecordRepository создает новый экземпляр HistoryRecordRepository с использованием предоставленного пула соединений к базе даннх.
func NewHistoryRecordRepository(pool *pgxpool.Pool) HistoryRecordRepository {
	return &repository.HistoryRecordPostgres{Pool: pool}
}
