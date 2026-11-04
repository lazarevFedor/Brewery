package class

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
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
		Name:     class.Name,
		ParentID: pgtype.Int8{Int64: *class.ParentID, Valid: true},
	}
	dbClass, err := repo.query.InsertClass(ctx, insertParams)
	if err != nil {
		return nil, fmt.Errorf("InsertClass: %w", err)
	}
	var parentID *int64
	if dbClass.ParentID.Valid {
		parentID = &dbClass.ParentID.Int64
	}

	return &entities.Class{
		ID:       dbClass.ID,
		Name:     dbClass.Name,
		ParentID: parentID,
	}, nil
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

func (repo *Repository) UpdateClassParent(ctx context.Context, id int64, parentID int64) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	updateParams := queries.UpdateParentClassParams{
		ParentID: pgtype.Int8{Int64: id, Valid: true},
		ID:       id,
	}
	dbClass, err := repo.query.UpdateParentClass(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("RenameClass: %w", err)
	}
	class := mapper.ToDomainClass(dbClass)
	return &class, nil
}

func (repo *Repository) UpdateClassName(ctx context.Context, id int64, name string) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	updateParams := queries.RenameClassParams{
		Name: name,
		ID:   id,
	}
	dbClass, err := repo.query.RenameClass(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("RenameClass: %w", err)
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
