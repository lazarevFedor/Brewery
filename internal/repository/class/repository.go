package class

import (
	"Brewery/internal/entities"
	mpr "Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var mapper = mpr.ClassMapperImpl{}

type Repository struct {
	Pool  *pgxpool.Pool
	query *queries.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Pool:  pool,
		query: queries.New(pool),
	}
}

func (repo *Repository) GetClasses(ctx context.Context) ([]entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	classes, err := repo.query.GetClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}
	return mapper.FromDBClasses(classes), nil
}

func (repo *Repository) InsertClass(ctx context.Context, class entities.Class) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	insertParams := queries.InsertClassParams{
		Name:     class.Name,
		ParentID: class.ParentID,
	}
	dbClass, err := repo.query.InsertClass(ctx, insertParams)
	if err != nil {
		return nil, fmt.Errorf("InsertClass: %w", err)
	}
	class = mapper.FromInsert(dbClass)
	return &class, nil
}
func (repo *Repository) GetClassByID(ctx context.Context, id int64) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	dbClass, err := repo.query.GetClassByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetClassByID: %w", err)
	}
	class := mapper.FromDBClass(dbClass)

	return &class, nil
}

func (repo *Repository) UpdateClass(ctx context.Context, id int64, updates entities.Class) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	updateParams := queries.UpdateClassParams{ID: id}

	if updates.ParentID == nil {
		updateParams.ParentID = nil
	}
	if updates.Name == "" {
		updateParams.Name = nil
	}

	dbClass, err := repo.query.UpdateClass(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("UpdateClass: %w", err)
	}
	class := mapper.FromDBClass(dbClass)
	return &class, nil
}

func (repo *Repository) DeleteClass(ctx context.Context, id int64) error {
	if repo.Pool == nil {
		return errors.New("db pool is empty")
	}

	if err := repo.query.DeleteClass(ctx, id); err != nil {
		return fmt.Errorf("DeleteClass: %w", err)
	}
	return nil
}

func (repo *Repository) GetClassByName(ctx context.Context, name string) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	dbClass, err := repo.query.GetClassByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("GetClassByName: %w", err)
	}
	class := mapper.FromDBClass(dbClass)
	return &class, nil
}
