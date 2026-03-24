package category

import (
	"Brewery/internal/entities"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableCategories = "product_categories"
	IDCol           = "id"
	nameCol         = "name"
	parentIDCol     = "parent_id"
)

type CategoryRepository interface {
	InsertCategory(ctx context.Context, category entities.ProductCategory) error
	DeleteCategoryByName(ctx context.Context, name string) error
}

type CategoryPostgres struct {
	pool *pgxpool.Pool
}

func NewCategoryPostgres(pool *pgxpool.Pool) *CategoryPostgres {
	return &CategoryPostgres{pool: pool}
}

func (r *CategoryPostgres) GetCategories(ctx context.Context) ([]entities.ProductCategory, error) {
	builder := sq.Select(IDCol, nameCol, parentIDCol).From(tableCategories)

	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}

	categories := make([]entities.ProductCategory, 0)
	for rows.Next() {
		ctg := entities.ProductCategory{}

		err = rows.Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
		if err != nil {
			return nil, fmt.Errorf("Scan: %w", err)
		}
		categories = append(categories, ctg)
	}
	return categories, nil
}

func (r *CategoryPostgres) InsertCategory(ctx context.Context, category entities.ProductCategory) error {
	data := map[string]interface{}{
		nameCol:     category.Name,
		parentIDCol: category.ParentID,
	}
	builder := sq.Insert(tableCategories).SetMap(data)
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("ToSql: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("Exec: %w", err)
	}
	return nil
}

func (r *CategoryPostgres) GetCategoryById(ctx context.Context, id int) (*entities.ProductCategory, error) {
	data := map[string]interface{}{
		IDCol: id,
	}
	builder := sq.Select(IDCol, nameCol, parentIDCol).Where(data)
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	var ctg entities.ProductCategory
	err = r.pool.QueryRow(ctx, query, args...).Scan(&ctg.ID, &ctg.Name, &ctg.ParentID)
	if err != nil {
		return nil, fmt.Errorf("QueryRow: %w", err)
	}

	return &ctg, nil
}

func (r *CategoryPostgres) UpdateCategory(ctx context.Context, id int, updates map[string]interface{}) error {
	builder := sq.Update(tableCategories).
		SetMap(updates).
		Where(sq.Eq{IDCol: id})
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("ToSql: %w", err)
	}

	result, err := r.pool.Exec(ctx, query, args...)
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

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete category: no such category")
	}

	return nil
}
