package repository

import (
	"Brewery/internal/entities"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	tableCategories = "product_categories"
	IDCol           = "id"
	nameCol         = "name"
	parentIDCol     = "parent_id"
)

type CategoryRepository interface {
	GetCategories(ctx context.Context) ([]entities.ProductCategory, error)
	InsertCategory(ctx context.Context, category entities.ProductCategory)(int, error)
	GetCategoryByID(ctx context.Context, id int) (*entities.ProductCategory, error)
	UpdateCategory(ctx context.Context, id int, updates map[string]any) error
	DeleteCategoryByID(ctx context.Context, id int) error
}

type CategoryPostgres struct {
	Pool *pgxpool.Pool
}

func NewCategoryPostgres(Pool *pgxpool.Pool) *CategoryPostgres {
	return &CategoryPostgres{Pool: Pool}
}

func (r *CategoryPostgres) GetCategories(ctx context.Context) ([]entities.ProductCategory, error) {
	builder := sq.Select(IDCol, nameCol, parentIDCol).From(tableCategories)

	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}

	categories := make([]entities.ProductCategory, 0)
	
	for rows.Next() {
		ctg := entities.ProductCategory{}

		var nullableInt pgtype.Int8
		err = rows.Scan(&ctg.ID, &ctg.Name, &nullableInt)
		if err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		if nullableInt.Valid {
			ctg.ParentID = int(nullableInt.Int64)
		} else {
			ctg.ParentID = 0
		}
		
		categories = append(categories, ctg)
	}
	return categories, nil
}

func (r *CategoryPostgres) InsertCategory(
	ctx context.Context, category entities.ProductCategory,
	)(int, error) {
	data := map[string]any{
		nameCol:     category.Name,
	}
	if category.ParentID != 0{
	data[parentIDCol] =  category.ParentID
	}

	builder := sq.Insert(tableCategories).SetMap(data).Suffix("RETURNING id")
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("ToSql: %w", err)
	}

	var newID int
	err = r.Pool.QueryRow(ctx, query, args...).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("Exec: %w", err)
	}

	return newID, nil

}

func (r *CategoryPostgres) GetCategoryByID(ctx context.Context, id int) (*entities.ProductCategory, error) {
	data := map[string]any{
		IDCol: id,
	}
	builder := sq.Select(IDCol, nameCol, parentIDCol).Where(data)
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	var ctg entities.ProductCategory
	err = r.Pool.QueryRow(ctx, query, args...).Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
	if err != nil {
		return nil, fmt.Errorf("QueryRow: %w", err)
	}

	return &ctg, nil
}

func (r *CategoryPostgres) UpdateCategory(ctx context.Context, id int, updates map[string]any) error {
	builder := sq.Update(tableCategories).
		SetMap(updates).
		Where(sq.Eq{IDCol: id})
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("ToSql: %w", err)
	}
	result, err := r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("QueryRow: %w", err)
	}

	if result.RowsAffected() == 0{
		return fmt.Errorf("failed to update category: no such category")
	}

	return nil
}

func (r *CategoryPostgres) DeleteCategoryByID(ctx context.Context, id int) error {
	conditions := map[string]interface{}{
		IDCol: id,
	}
	builder := sq.Delete(tableCategories).Where(conditions)
	psql := builder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("ToSql: %w", err)
	}

	result, err := r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete category: no such category")
	}

	return nil
}
