package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ParameterRepository определяет интерфейс для работы с параметрами, включая числовые и перечисляемые параметры.
type ParameterRepository interface {
	// InsertRangeParameter добавляет новый числовой параметр в базу данных и возвращает его.
	InsertRangeParameter(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error)

	// UpdateRangeParameter обновляет существующий числовой параметр в базе данных и возвращает его.
	UpdateRangeParameter(ctx context.Context, id int64, updates map[string]any) (*entities.RangeParameter, error)

	// DeleteRangeParameter удаляет числовой параметр из базы данных и возвращает его.
	DeleteRangeParameter(ctx context.Context, id int64) (*entities.RangeParameter, error)

	// GetParameters извлекает все числовые и перечисляемые параметры из базы данных и возвращает их.
	GetParameters(ctx context.Context, classID int64, parameterType int) ([]entities.RangeParameter, []entities.EnumParameter, error)

	// ApplyParameters применяет числовые и перечисляемые параметры к категории и возвращает результат применения.
	ApplyParameters(ctx context.Context, classID int64, rangeParameters, enumParameters []int) (int, error)

	// ExcludeParameters исключает указанные числовые и перечисляемые параметры из категории и возвращает количество исключенных параметров.
	ExcludeParameters(ctx context.Context, classID int64, rangeParameters, enumParameters []int) (int, error)
}

// NewParameterRepository создает новый экземпляр ParameterRepository с предоставленным пулом соединений к базе данных.
func NewParameterRepository(pool *pgxpool.Pool) ParameterRepository {
	return &repository.ParameterPostgres{Pool: pool}
}

// ParametersService - реализация ParametersService.
type ParametersService struct {
	ParamRepo ParameterRepository
	EnumRepo  EnumRepository
}

// allowedParametersToUpdate содержит список полей, которые можно обновлять в числовом параметре.
var allowedParametersToUpdate = map[string]struct{}{
	"min_val":     {},
	"max_val":     {},
	"inheritable": {},
	"unit":        {},
}

// CreateRange создает числовой параметр.
func (s *ParametersService) CreateRange(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if param == nil {
		return nil, errors.New("parameter is nil")
	}

	if param.FieldName == "" {
		return nil, errors.New("field_name is required")
	}
	if param.MinValue > param.MaxValue {
		return nil, errors.New("min_val must be less than or equal to max_val")
	}
	if param.MinValue < 0 {
		return nil, errors.New("min_val must be greater or equal to zero")
	}

	created, err := s.ParamRepo.InsertRangeParameter(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create range parameter: %w", err)
	}

	return created, nil
}

// UpdateRange обновляет числовой параметр. Единственные поля, которые можно обновлять - min_val, max_val и inheritable.
func (s *ParametersService) UpdateRange(ctx context.Context, id int64, updates map[string]any) (*entities.RangeParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, errors.New("invalid parameter ID")
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	for k := range updates {
		if _, ok := allowedParametersToUpdate[k]; !ok {
			return nil, fmt.Errorf("unknown parameter: %s", k)
		}
	}

	updated, err := s.ParamRepo.UpdateRangeParameter(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update range parameter: %w", err)
	}

	return updated, nil
}

// DeleteRange удаляет числовой параметр.
func (s *ParametersService) DeleteRange(ctx context.Context, id int64) (*entities.RangeParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, errors.New("invalid parameter ID")
	}

	deleted, err := s.ParamRepo.DeleteRangeParameter(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete range parameter: %w", err)
	}

	return deleted, nil
}

// ListParameters возвращает все параметры, применимые к классу.
func (s *ParametersService) ListParameters(ctx context.Context, classID int64, parameterType int) ([]entities.RangeParameter, []entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("request cancelled: %w", err)
	}

	if classID == 0 {
		return nil, nil, errors.New("invalid class ID")
	}

	numeric, enum, err := s.ParamRepo.GetParameters(ctx, int64(classID), parameterType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get parameters: %w", err)
	}

	return numeric, enum, nil
}

// ApplyToClass применяет указанные параметры к классу. Возвращает количество добавленных параметров.
func (s *ParametersService) ApplyToClass(ctx context.Context, classID int64, rangeIDs, enumIDs []int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if classID == 0 {
		return 0, errors.New("invalid class ID")
	}

	if len(rangeIDs) == 0 && len(enumIDs) == 0 {
		return 0, errors.New("no parameters to apply")
	}

	added, err := s.ParamRepo.ApplyParameters(ctx, classID, rangeIDs, enumIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to apply parameters: %w", err)
	}

	return added, nil
}

func (s *ParametersService) ExcludeFromClass(ctx context.Context, classID int64, rangeIDs, enumIDs []int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if classID == 0 {
		return 0, errors.New("invalid class ID")
	}

	if len(rangeIDs) == 0 && len(enumIDs) == 0 {
		return 0, errors.New("no parameters to exclude")
	}

	excluded, err := s.ParamRepo.ExcludeParameters(ctx, classID, rangeIDs, enumIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to exclude parameters: %w", err)
	}

	return excluded, nil
}
