package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"Brewery/pkg/logger"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	tableCategories = "product_categories"
	IDCol           = "id"
	nameCol         = "name"
	parentIDCol     = "parent_id"
)

type CategoryRepository interface {
	GetCategories(ctx context.Context) ([]entities.ProductCategory, error)
	InsertCategory(ctx context.Context, category entities.ProductCategory) (int, error)
	GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error)
	UpdateCategory(ctx context.Context, id uint, updates map[string]any) error
	DeleteCategoryByID(ctx context.Context, id uint) error
}

type CategoryPostgres struct {
	Pool *pgxpool.Pool
}

func NewCategoryPostgres(pool *pgxpool.Pool) *CategoryPostgres {
	return &CategoryPostgres{Pool: pool}
}

func (r *CategoryPostgres) GetCategories(ctx context.Context) ([]entities.ProductCategory, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.FullCategorySelect()

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Query", err)
	}

	categories := make([]entities.ProductCategory, 0)

	for rows.Next() {
		ctg := entities.ProductCategory{}

		var nullableInt pgtype.Int8
		err = rows.Scan(&ctg.ID, &ctg.Name, &nullableInt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "Scan", err)
		}
		if nullableInt.Valid {
			ctg.ParentID = int(nullableInt.Int64)
		} else {
			ctg.ParentID = 0
		}

		categories = append(categories, ctg)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return categories, nil
}

func (r *CategoryPostgres) InsertCategory(
	ctx context.Context, category entities.ProductCategory,
) (int, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.CategoryInsert(category)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var newID int
	err = tx.QueryRow(ctx, query, args...).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Exec", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return newID, nil
}

func (r *CategoryPostgres) GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.SelectCategoryByID(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var ctg entities.ProductCategory
	err = tx.QueryRow(ctx, query, args...).Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "QueryRow", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ctg, nil
}

func (r *CategoryPostgres) UpdateCategory(ctx context.Context, id uint, updates map[string]any) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.UpdateCategory(id, updates)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}
	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "QueryRow", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("failed to update category: no such category")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *CategoryPostgres) DeleteCategoryByID(ctx context.Context, id uint) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)

		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, fmt.Sprintf("%s InsertBeer: Rollback:", "InsertBeer"), zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.DeleteCategory(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "Exec", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("failed to delete category: no such category")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
