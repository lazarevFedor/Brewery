// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/apperrors"
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

type CategoryRepository interface {
	// GetCategories получает все категории из базы данных.
	GetCategories(ctx context.Context) ([]entities.ProductCategory, error)

	// InsertCategory вставляет новую категорию в базу данных.
	InsertCategory(ctx context.Context, tx pgx.Tx, category entities.ProductCategory) (uint, error)

	// GetCategoryByID получает категорию по её ID.
	GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error)

	// UpdateCategory обновляет поля категории по её ID.
	UpdateCategory(ctx context.Context, id uint, updates map[string]any) error

	// DeleteCategoryByID удаляет категорию по её ID.
	DeleteCategoryByID(ctx context.Context, id uint) error

	// GetCategoryID получает ID категории по её имени.
	GetCategoryID(ctx context.Context, tx pgx.Tx, ctgName string) (uint, error)
}

// CategoryPostgres реализует интерфейс CategoryRepository для работы с категориями в базе данных PostgreSQL.
// Если пул равен nil, методы будут возвращать ошибку при попытке доступа к базе данных.
type CategoryPostgres struct {
	Pool *pgxpool.Pool
}

// NewCategoryRepository создает новый экземпляр CategoryRepository с переданным пулом соединений.
func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
	return &CategoryPostgres{Pool: pool}
}

// NewCategoryPostgres создает новый экземпляр CategoryPostgres с переданным пулом соединений.
// Если пул равен nil, методы будут возвращать ошибку при попытке доступа к базе данных.
func NewCategoryPostgres(pool *pgxpool.Pool) *CategoryPostgres {
	return &CategoryPostgres{Pool: pool}
}

// GetCategories получает все категории из базы данных. Если категорий нет, возвращает пустой срез и nil.
func (r *CategoryPostgres) GetCategories(ctx context.Context) ([]entities.ProductCategory, error) {
	if r.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.FullCategorySelect()
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build FullCategorySelect query: %w", err))
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute FullCategorySelect query: %w", err))
	}

	categories := make([]entities.ProductCategory, 0)

	for rows.Next() {
		ctg := entities.ProductCategory{}

		err = rows.Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
		if err != nil {
			return nil, apperrors.Internal(fmt.Errorf("scan category's row into ProductCategory struct: %w", err))
		}
		categories = append(categories, ctg)
	}

	return categories, nil
}

// InsertCategory вставляет новую категорию в базу данных. Если категория с таким именем уже существует,
// возвращает ошибку. Если ParentID не равен 0, проверяет, что родительская категория существует.
// Возвращает ID новой категории.
func (r *CategoryPostgres) InsertCategory(ctx context.Context, tx pgx.Tx, category entities.ProductCategory) (uint, error) {
	if r.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.CategoryInsert(category)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("build CategoryInsert query: %w", err))
	}

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, args...)
	} else {
		row = r.Pool.QueryRow(ctx, query, args...)
	}

	var categoryID uint
	if err = row.Scan(&categoryID); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("scan category's row: %w", err))
	}

	return categoryID, nil
}

// GetCategoryByID получает категорию по её ID. Если категория не найдена, возвращает nil и ошибку.
func (r *CategoryPostgres) GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error) {
	if r.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectCategoryByID(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build SelectCategoryByID query: %w", err))
	}

	var ctg entities.ProductCategory
	err = r.Pool.QueryRow(ctx, query, args...).Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute SelectCategoryByID query and scan ProductCategory: %w", err))
	}

	return &ctg, nil
}

// UpdateCategory обновляет поля категории по её ID. Если категория не найдена, возвращает ошибку. Если updates пустой, ничего не обновляет.
func (r *CategoryPostgres) UpdateCategory(ctx context.Context, id uint, updates map[string]any) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("begin UpdateCategory transaction: %w", err))
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "roll back UpdateCategory transaction", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.UpdateCategory(id, updates)
	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("build UpdateCategory query: %w", err))
	}
	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("execute UpdateCategory query: %w", err))
	}

	if result.RowsAffected() == 0 {
		return apperrors.BadRequest("no such category", errors.New("failed to update category: no such category"))
	}

	err = tx.Commit(ctx)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("commit UpdateCategory transaction: %w", err))
	}
	return nil
}

// DeleteCategoryByID удаляет категорию по её ID. Если категория имеет дочерние категории, они будут перемещены к родителю удаляемой категории. Нельзя удалить корневую категорию.
func (r *CategoryPostgres) DeleteCategoryByID(ctx context.Context, id uint) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("begin DeleteCategoryByID transaction: %w", err))
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)

		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "roll back DeleteCategoryByID transaction", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	category, err := r.GetCategoryByID(ctx, id)
	if err != nil {
		return err
	}

	if category.ParentID == 0 {
		return apperrors.BadRequest("cannot delete root category", errors.New("cannot delete root category"))
	}

	childrenPsql := queries.SelectChildrenCategories(id)
	query, args, err := childrenPsql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("build SelectChildrenCategories query: %w", err))
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("execute SelectChildrenCategories query: %w", err))
	}

	for rows.Next() {
		childID := 0
		err = rows.Scan(&childID)
		if err != nil {
			return apperrors.Internal(fmt.Errorf("scan category's children: %w", err))
		}
		if childID != 0 {
			err = r.UpdateCategory(ctx, uint(childID),
				map[string]any{
					"parent_id": category.ParentID,
				})
			if err != nil {
				return err
			}
		}
	}

	deletePsql := queries.DeleteCategory(id)
	query, args, err = deletePsql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("build DeleteCategory query: %w", err))
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute DeleteCategory query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.BadRequest("no such category", errors.New("failed to delete category: no such category"))
	}

	err = tx.Commit(ctx)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("commit DeleteCategoryByID transaction: %w", err))
	}
	return nil
}

// GetCategoryID получает ID категории по её имени. Если категория не найдена, возвращает 0 и ошибку.
func (r *CategoryPostgres) GetCategoryID(ctx context.Context, tx pgx.Tx, ctgName string) (uint, error) {
	if r.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectCategoryByName(ctgName)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("build SelectCategoryByName query: %w", err))
	}

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, args...)
	} else {
		row = r.Pool.QueryRow(ctx, query, args...)
	}

	var categoryID uint
	if err = row.Scan(&categoryID); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, apperrors.Internal(fmt.Errorf("scan category ID from row: %w", err))
		}
	}

	return categoryID, nil
}
