package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RangeParametersRepository определяет операции хранения числовых параметров.
type RangeParametersRepository interface {
	// InsertRangeParameter добавляет новый числовой параметр в базу данных и возвращает его.
	InsertRangeParameter(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error)

	// GetRangeParameters возвращает все числовые параметры или параметры с указанным именем поля.
	GetRangeParameters(ctx context.Context, fieldName string) ([]entities.RangeParameter, error)

	// UpdateRangeParameter обновляет существующий числовой параметр в базе данных и возвращает его.
	UpdateRangeParameter(ctx context.Context, id int64, updates map[string]any) (*entities.RangeParameter, error)

	// DeleteRangeParameter удаляет числовой параметр из базы данных и возвращает его.
	DeleteRangeParameter(ctx context.Context, id int64) (*entities.RangeParameter, error)
}

// NewRangeParametersRepository создает репозиторий числовых параметров.
func NewRangeParametersRepository(pool *pgxpool.Pool) RangeParametersRepository {
	return repository.NewRangeParametersPostgres(pool)
}

// RangeParametersService реализует бизнес-логику числовых параметров.
type RangeParametersService struct {
	repo RangeParametersRepository
}

// NewRangeParametersService создает сервис числовых параметров.
func NewRangeParametersService(repo RangeParametersRepository) *RangeParametersService {
	return &RangeParametersService{repo: repo}
}

// allowedRangeParameterUpdates содержит список полей, которые можно обновлять в числовом параметре.
var allowedRangeParameterUpdates = map[string]struct{}{
	"min_val":     {},
	"max_val":     {},
	"inheritable": {},
	"unit":        {},
}

// CreateRangeParameter создает числовой параметр.
func (s *RangeParametersService) CreateRangeParameter(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error) {
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

	created, err := s.repo.InsertRangeParameter(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create range parameter: %w", err)
	}

	return created, nil
}

// ListRangeParameters возвращает числовые параметры; пустое имя поля отключает фильтрацию.
func (s *RangeParametersService) ListRangeParameters(ctx context.Context, fieldName string) ([]entities.RangeParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	parameters, err := s.repo.GetRangeParameters(ctx, fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to get range parameters: %w", err)
	}

	return parameters, nil
}

// UpdateRangeParameter обновляет числовой параметр. Обновлять можно min_val, max_val, inheritable и unit.
func (s *RangeParametersService) UpdateRangeParameter(ctx context.Context, id int64, updates map[string]any) (*entities.RangeParameter, error) {
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
		if _, ok := allowedRangeParameterUpdates[k]; !ok {
			return nil, fmt.Errorf("unknown parameter: %s", k)
		}
	}

	updated, err := s.repo.UpdateRangeParameter(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update range parameter: %w", err)
	}

	return updated, nil
}

// DeleteRangeParameter удаляет числовой параметр.
func (s *RangeParametersService) DeleteRangeParameter(ctx context.Context, id int64) (*entities.RangeParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, errors.New("invalid parameter ID")
	}

	deleted, err := s.repo.DeleteRangeParameter(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete range parameter: %w", err)
	}

	return deleted, nil
}
