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

// ClassPostgres реализует интерфейс ClassRepository для работы с классами в базе данных PostgreSQL.
// Если пул равен nil, методы будут возвращать ошибку при попытке доступа к базе данных.
type ClassPostgres struct {
	Pool *pgxpool.Pool
}

// NewClassPostgres создает новый экземпляр ClassPostgres с переданным пулом соединений.
// Если пул равен nil, методы будут возвращать ошибку при попытке доступа к базе данных.
func NewClassPostgres(pool *pgxpool.Pool) *ClassPostgres {
	return &ClassPostgres{Pool: pool}
}

// GetClasses получает все классы из базы данных. Если классов нет, возвращает пустой срез и nil.
func (r *ClassPostgres) GetClasses(ctx context.Context) ([]entities.Class, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.FullCategorySelect()
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Query", err)
	}

	categories := make([]entities.Class, 0)

	for rows.Next() {
		ctg := entities.Class{}

		err = rows.Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "Scan", err)
		}
		categories = append(categories, ctg)
	}

	return categories, nil
}

// InsertClass вставляет новую категорию в базу данных. Если категория с таким именем уже существует,
// возвращает ошибку. Если ParentID не равен 0, проверяет, что родительская категория существует.
// Возвращает ID новой категории.
func (r *ClassPostgres) InsertClass(ctx context.Context, tx pgx.Tx, class entities.Class) (uint, error) {
	if r.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.CategoryInsert(class)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("%s: %w", "ToSql", err))
	}

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, args...)
	} else {
		row = r.Pool.QueryRow(ctx, query, args...)
	}

	var categoryID uint
	if err = row.Scan(&categoryID); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("category QueryRow: %w", err))
	}

	return categoryID, nil
}

// GetClassByID получает класс по его ID. Если класс не найден, возвращает nil и ошибку.
func (r *ClassPostgres) GetClassByID(ctx context.Context, id uint) (*entities.Class, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectCategoryByID(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var ctg entities.Class
	err = r.Pool.QueryRow(ctx, query, args...).Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "QueryRow", err)
	}

	return &ctg, nil
}

// UpdateClass обновляет поля класса по его ID. Если класс не найден, возвращает ошибку. Если updates пустой, ничего не обновляет.
func (r *ClassPostgres) UpdateClass(ctx context.Context, id uint, updates map[string]any) error {
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

// DeleteClassByID удаляет класс по его ID. Если класс имеет дочерние классы, они будут перемещены к родителю удаляемого класса. Нельзя удалить корневой класс.
func (r *ClassPostgres) DeleteClassByID(ctx context.Context, id uint) error {
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

	class, err := r.GetClassByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get class by id: %w", err)
	}

	if class.ParentID == 0 {
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
			err = r.UpdateClass(ctx, uint(childID),
				map[string]any{
					"parent_id": class.ParentID,
				})
			if err != nil {
				return fmt.Errorf("failed to update class: %w", err)
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

// GetClassID получает ID класса по его имени. Если класс не найден, возвращает 0 и ошибку.
func (r *ClassPostgres) GetClassID(ctx context.Context, tx pgx.Tx, ctgName string) (uint, error) {
	if r.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectCategoryByName(ctgName)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
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
			return 0, apperrors.Internal(fmt.Errorf("QueryRow: %w", err))
		}
	}

	return categoryID, nil
}
