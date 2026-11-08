package history

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	mpr "Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	Pool   *pgxpool.Pool
	query  *queries.Queries
	mapper mpr.HistoryMapperImpl
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Pool:  pool,
		query: queries.New(pool),
	}
}

func (repo *Repository) InsertChangeRecord(ctx context.Context, record entities.HistoryRecord) (*entities.HistoryRecord, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	params := queries.InsertChangeParams{
		ProductID: record.ProductID,
		Changes:   record.Changes,
	}
	dbRecord, err := repo.query.InsertChange(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "change_history_product_id_fkey" {
			return nil, apperrors.NotFound(fmt.Sprintf("Product with id %d not found", record.ProductID), err)
		}
		return nil, fmt.Errorf("InsertChange: %w", err)
	}
	record = repo.mapper.FromDBChangeRecord(dbRecord)
	return &record, nil
}

func (repo *Repository) GetHistoryByProduct(
	ctx context.Context, productID int64,
	start, end time.Time,
	limit, offset int32,
) ([]entities.HistoryRecord, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	builder := squirrel.Select("id", "name", "amount", "created_at").
		From("change_history").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"product_id": productID})

	if !start.IsZero() {
		builder = builder.Where(squirrel.GtOrEq{"created_at": start})
	}
	if !end.IsZero() {
		builder = builder.Where(squirrel.LtOrEq{"created_at": start})
	}
	if limit != 0 {
		builder = builder.Limit(uint64(limit))
	}
	if offset != 0 {
		builder = builder.Limit(uint64(offset))
	}
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	// 2. Выполняем собранный запрос через pgxpool
	rows, err := repo.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbHistory []queries.ChangeHistory // Можно переиспользовать структуру от sqlc!
	for rows.Next() {
		var p queries.ChangeHistory
		if err := rows.Scan(&p.ID, &p.ProductID, &p.Changes, &p.CreatedAt); err != nil {
			return nil, err
		}
		dbHistory = append(dbHistory, p)
	}
	return repo.mapper.FromDBChangeHistory(dbHistory), nil
}

func (repo *Repository) GetHistoryRecordByID(ctx context.Context, id int64) ([]entities.HistoryRecord, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	dbRecord, err := repo.query.GetHistoryRecordByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("InsertChange: %w", err)
	}
	history := repo.mapper.FromDBChangeHistory(dbRecord)
	return history, nil
}
