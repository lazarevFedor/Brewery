package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/history"
	"context"
	"fmt"
	"strings"
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
	return history.NewRepository(pool)
}

type HistoryService struct {
	repo HistoryRecordRepository
}

// NewHistoryService creates history service.
func NewHistoryService(repo HistoryRecordRepository) *HistoryService {
	return &HistoryService{repo: repo}
}

func (hs *HistoryService) CreateChangeRecord(ctx context.Context, record *entities.HistoryRecord) (*entities.HistoryRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if record == nil {
		return nil, apperrors.BadRequest("record cannot be nil", nil)
	}
	if record.ProductID <= 0 {
		return nil, apperrors.BadRequest("invalid product id", nil)
	}
	if strings.TrimSpace(record.Changes) == "" {
		return nil, apperrors.BadRequest("changes is required", nil)
	}

	record.CreatedAt = time.Now()

	record, err := hs.repo.InsertChangeRecord(ctx, *record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (hs *HistoryService) GetHistoryByProduct(
	ctx context.Context, productID int64,
	start, end time.Time,
	limit, offset int32,
) ([]entities.HistoryRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if productID <= 0 {
		return nil, apperrors.BadRequest("invalid product id", nil)
	}
	if limit < 0 || offset < 0 {
		return nil, apperrors.BadRequest("invalid pagination", nil)
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return nil, apperrors.BadRequest("start date cannot be after end date", nil)
	}

	records, err := hs.repo.GetHistoryByProduct(ctx, productID, start, end, limit, offset)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (hs *HistoryService) GetHistoryRecordByID(ctx context.Context, id int64) ([]entities.HistoryRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return nil, apperrors.BadRequest("invalid history record id", nil)
	}

	historyRecords, err := hs.repo.GetHistoryRecordByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return historyRecords, nil
}
