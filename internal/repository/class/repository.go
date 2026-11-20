package class

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository"
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

func NewClassRepository(pool *pgxpool.Pool) *Repository {
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
	tx, err := repo.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if class.ParentID != nil {
		if lockErr := repository.LockParameterDefinitions(ctx, tx); lockErr != nil {
			return nil, lockErr
		}
	}
	txQueries := repo.query.WithTx(tx)
	insertParams := queries.InsertClassParams{
		Name:     class.Name,
		ParentID: class.ParentID,
	}
	dbClass, err := txQueries.InsertClass(ctx, insertParams)
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
	if class.ParentID != nil {
		class, err = inheritClassParameters(ctx, txQueries, class.ID, *class.ParentID)
		if err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
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
	tx, err := repo.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txQueries := repo.query.WithTx(tx)
	updateParams := queries.UpdateClassParams{ID: id}

	if updates.ParentID != nil {
		if lockErr := repository.LockParameterDefinitions(ctx, tx); lockErr != nil {
			return nil, lockErr
		}
		previous, readErr := txQueries.GetClassByID(ctx, id)
		if errors.Is(readErr, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("Class not found", readErr)
		}
		if readErr != nil {
			return nil, readErr
		}
		if previous.ParentID == nil || *previous.ParentID != *updates.ParentID {
			cycle, ancestorErr := txQueries.ClassMoveCreatesCycle(ctx, queries.ClassMoveCreatesCycleParams{
				ClassID: *updates.ParentID, AncestorID: id,
			})
			if ancestorErr != nil {
				return nil, ancestorErr
			}
			if cycle {
				return nil, apperrors.BadRequest("class parent creates a cycle", nil)
			}
			updateParams.ParentID = updates.ParentID
		}
	}
	if updates.Name != "" {
		updateParams.Name = &updates.Name
	}

	dbClass, err := txQueries.UpdateClass(ctx, updateParams)
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
	if updateParams.ParentID != nil {
		class, err = inheritClassParameters(ctx, txQueries, class.ID, *updateParams.ParentID)
		if err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &class, nil
}

func inheritClassParameters(ctx context.Context, txQueries *queries.Queries, classID, parentID int64) (entities.Class, error) {
	err := txQueries.InheritClassBranchParameters(ctx, queries.InheritClassBranchParametersParams{ClassID: classID, ParentID: parentID})
	if err != nil {
		return entities.Class{}, fmt.Errorf("inherit class parameters: %w", repository.ParameterFieldNameError(err))
	}
	row, err := txQueries.GetClassByID(ctx, classID)
	if err != nil {
		return entities.Class{}, err
	}
	return mapper.FromDBClass(row), nil
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
