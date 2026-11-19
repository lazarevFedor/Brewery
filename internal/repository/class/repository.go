package class

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	mpr "Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"Brewery/pkg/logger"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var mapper = mpr.ClassMapperImpl{}
var log, _ = logger.GetLoggerFromCtx(context.TODO())

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
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch {
			case pgErr.Code == "23505" && pgErr.ConstraintName == "idx_classes_single_root":
				return nil, apperrors.Conflict("root class already exists", err)
			case pgErr.Code == "23505" && pgErr.ConstraintName == "classes_name_key":
				return nil, apperrors.Conflict("class name already exists", err)
			case pgErr.Code == "23503" && pgErr.ConstraintName == "classes_parent_id_fkey":
				return nil, apperrors.BadRequest("parent class not found", err)
			}
		}
		return nil, fmt.Errorf("InsertClass: %w", err)
	}
	class = mapper.FromInsert(dbClass)
	return &class, nil
}
func (repo *Repository) GetClassByID(ctx context.Context, id int64) (*entities.Class, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	log.Debug(ctx, "class id", zap.Int64("id", id))
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

	if updates.ParentID != nil {
		updateParams.ParentID = updates.ParentID
	}
	if updates.Name != "" {
		updateParams.Name = &updates.Name
	}

	dbClass, err := repo.query.UpdateClass(ctx, updateParams)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.NotFound("Class not found", err)
	}
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch {
			case pgErr.Code == "23505" && pgErr.ConstraintName == "classes_name_key":
				return nil, apperrors.Conflict("class name already exists", err)
			case pgErr.Code == "23503" && pgErr.ConstraintName == "classes_parent_id_fkey":
				return nil, apperrors.NotFound("Parent class not found", err)
			}
		}
		return nil, fmt.Errorf("UpdateClass: %w", err)
	}

	class := mapper.FromDBClass(dbClass)
	return &class, nil
}

func (repo *Repository) DeleteClass(ctx context.Context, id int64) error {
	if repo.Pool == nil {
		return errors.New("db pool is empty")
	}

	deleted, err := repo.query.DeleteClass(ctx, id)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && (pgErr.Code == "23503" || pgErr.Code == "23001") {
			return apperrors.Conflict("class is referenced by another record", err)
		}
		return fmt.Errorf("DeleteClass: %w", err)
	}
	if deleted == 0 {
		return apperrors.NotFound("Class not found", nil)
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
