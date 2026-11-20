package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ClassParametersRepository определяет операции с параметрами классов изделий.
type ClassParametersRepository interface {
	// GetParameters извлекает все числовые и перечисляемые параметры из базы данных и возвращает их.
	GetParameters(ctx context.Context, classID int64, parameterType int) ([]entities.RangeParameter, []entities.EnumParameter, error)

	// ApplyParameters применяет числовые и перечисляемые параметры к классу и возвращает результат применения.
	ApplyParameters(ctx context.Context, classID int64, rangeParameters, enumParameters []int64) (int, error)

	// ExcludeParameters исключает указанные числовые и перечисляемые параметры из класса и возвращает число фактически исключенных параметров.
	ExcludeParameters(ctx context.Context, classID int64, rangeParameters, enumParameters []int64) (int, error)
}

// NewClassParametersRepository создает репозиторий параметров классов.
func NewClassParametersRepository(pool *pgxpool.Pool) ClassParametersRepository {
	return repository.NewClassParametersPostgres(pool)
}

// ListParameters возвращает параметры класса; classID = 0 отключает фильтрацию по классу.
func (c *ClassService) ListParameters(ctx context.Context, classID int64, parameterType int) ([]entities.RangeParameter, []entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("request cancelled: %w", err)
	}

	if classID < 0 {
		return nil, nil, apperrors.BadRequest("invalid class ID", nil)
	}

	rangeParameters, enumParameters, err := c.parametersRepo.GetParameters(ctx, classID, parameterType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get parameters: %w", err)
	}

	return rangeParameters, enumParameters, nil
}

// ApplyParametersToClass применяет указанные параметры к классу. Возвращает количество добавленных параметров.
func (c *ClassService) ApplyParametersToClass(ctx context.Context, classID int64, rangeIDs, enumIDs []int64) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if classID <= 0 {
		return 0, apperrors.BadRequest("invalid class ID", nil)
	}

	if len(rangeIDs) == 0 && len(enumIDs) == 0 {
		return 0, apperrors.BadRequest("no parameters to apply", nil)
	}

	added, err := c.parametersRepo.ApplyParameters(ctx, classID, rangeIDs, enumIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to apply parameters: %w", err)
	}

	return added, nil
}

// ExcludeParametersFromClass исключает указанные параметры из класса.
func (c *ClassService) ExcludeParametersFromClass(ctx context.Context, classID int64, rangeIDs, enumIDs []int64) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if classID <= 0 {
		return 0, apperrors.BadRequest("invalid class ID", nil)
	}

	if len(rangeIDs) == 0 && len(enumIDs) == 0 {
		return 0, apperrors.BadRequest("no parameters to exclude", nil)
	}

	excluded, err := c.parametersRepo.ExcludeParameters(ctx, classID, rangeIDs, enumIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to exclude parameters: %w", err)
	}

	return excluded, nil
}
