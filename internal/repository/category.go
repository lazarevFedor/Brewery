package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"Brewery/pkg/logger"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	InsertCategory(ctx context.Context, category entities.ProductCategory) (uint, error)
	GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error)
	UpdateCategory(ctx context.Context, id uint, updates map[string]any) error
	DeleteCategoryByID(ctx context.Context, id uint) error
	GetCategoryID(ctx context.Context, ctgName string) (uint, error)
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
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.FullCategorySelect()
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Query", err)
	}

	categories := make([]entities.ProductCategory, 0)

	for rows.Next() {
		ctg := entities.ProductCategory{}

		err = rows.Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "Scan", err)
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
) (uint, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.CategoryInsert(category)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var newID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Exec", err)
	}

	if newID == 0 {
		return 0, errors.New("zero id")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return newID, nil
}

func (r *CategoryPostgres) GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error) {
	psql := queries.SelectCategoryByID(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var ctg entities.ProductCategory
	err = r.Pool.QueryRow(ctx, query, args...).Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "QueryRow", err)
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
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
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
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, fmt.Sprintf("%s InsertBeer: Rollback:", "InsertBeer"), zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	category, err := r.GetCategoryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get category by id: %w", err)
	}

	if category.ParentID == 0 {
		return errors.New("cannot delete root category")
	}

	childrenPsql := queries.SelectChildrenCategories(id)
	query, args, err := childrenPsql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "Query", err)
	}

	for rows.Next() {
		childID := 0
		err = rows.Scan(&childID)
		if err != nil {
			return fmt.Errorf("children scanning: %w", err)
		}
		if childID != 0 {
			err = r.UpdateCategory(ctx, uint(childID),
				map[string]any{
					"parent_id": category.ParentID,
				})
			if err != nil {
				return fmt.Errorf("failed to update category: %w", err)
			}
		}
	}

	deletePsql := queries.DeleteCategory(id)
	query, args, err = deletePsql.ToSql()
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

func (r *CategoryPostgres) GetCategoryID(ctx context.Context, ctgName string) (uint, error) {
	var categoryID uint
	psql := queries.SelectCategoryByName(ctgName)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Query", err)
	}

	if rows.Next() {
		err = rows.Scan(&categoryID)
		if err != nil {
			return 0, fmt.Errorf("scan: %w", err)
		}
	}
	rows.Close()

	return categoryID, nil
}
