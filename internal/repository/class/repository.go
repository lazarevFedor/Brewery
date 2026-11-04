package class

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
	return mapper.ToDomainClasses(classes), nil
}

func (repo *Repository) InsertClass(ctx context.Context, class entities.Class) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	insertParams := queries.InsertClassParams{
		Name:     *class.Name,
		ParentID: class.ParentID,
	}
	dbClass, err := repo.query.InsertClass(ctx, insertParams)
	if err != nil {
		return nil, fmt.Errorf("InsertClass: %w", err)
	}
	class = mapper.BaseToDomainClass(dbClass)
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
	class := mapper.ToDomainClass(dbClass)

	return &class, nil
}

func (repo *Repository) UpdateClass(ctx context.Context, classUpdates entities.Class) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	updateParams := queries.UpdateClassParams{
		Name:                classUpdates.Name,
		ParentID:            classUpdates.ParentID,
		ID:                  classUpdates.ID,
		NumericParameterIds: classUpdates.NumParamIDs,
		EnumParameterIds:    classUpdates.EnumParamIDs,
	}
	dbClass, err := repo.query.UpdateClass(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("UpdateClass: %w", err)
	}
	class := mapper.ToDomainClass(dbClass)
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
	class := mapper.ToDomainClass(dbClass)
	return &class, nil
}
