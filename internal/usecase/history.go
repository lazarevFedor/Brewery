package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/history"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HistoryRecordRepository определяет интерфейс для работы с записями истории в базе данных.
type HistoryRecordRepository interface {
	InsertChangeRecord(ctx context.Context, record entities.HistoryRecord) (*entities.HistoryRecord, error)
	GetHistoryByProduct(
		ctx context.Context, productID int64,
		start, end time.Time,
		limit, offset int32,
	) ([]entities.HistoryRecord, error)
	GetHistoryRecordByID(ctx context.Context, id int64) ([]entities.HistoryRecord, error)
}

// NewHistoryRecordRepository создает новый экземпляр HistoryRecordRepository с использованием предоставленного пула соединений к базе даннх.
func NewHistoryRecordRepository(pool *pgxpool.Pool) HistoryRecordRepository {
	return &history.Repository{Pool: pool}
}
