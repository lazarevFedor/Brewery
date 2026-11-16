// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"fmt"
	"maps"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnumValuesRepository определяет операции хранения значений перечисляемых параметров.
type EnumValuesRepository interface {
	// InsertEnumValue сохраняет новую сущность EnumValue в хранилище.
	InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error)

	// UpdateEnumValue обновляет сущность EnumValue в хранилище.
	UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumValueByID удаляет сущность EnumValue из хранилища.
	DeleteEnumValueByID(ctx context.Context, id uint) error

	// GetEnumValues получает список сущностей EnumValue по заданным имени поля и типу значения.
	GetEnumValues(ctx context.Context, field string, valueType entities.EnumType) ([]entities.EnumValue, error)

	// GetEnumValuesByParameterID получает список сущностей EnumValue по ID перечисляемого параметра.
	GetEnumValuesByParameterID(ctx context.Context, parameterID uint) ([]entities.EnumValue, error)
}

// NewEnumValuesRepository создает репозиторий значений перечислений.
func NewEnumValuesRepository(pool *pgxpool.Pool) EnumValuesRepository {
	return repository.NewEnumValuesPostgres(pool)
}

// EnumValuesService реализует бизнес-логику значений перечисляемых параметров.
type EnumValuesService struct {
	repo EnumValuesRepository
}

// NewEnumValuesService создает сервис значений перечислений.
func NewEnumValuesService(repo EnumValuesRepository) *EnumValuesService {
	return &EnumValuesService{repo: repo}
}

// CreateEnumValue создает новое значение перечисления и возвращает его ID.
func (s *EnumValuesService) CreateEnumValue(ctx context.Context, value entities.EnumValue) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if value.EnumClassID == 0 {
		return 0, fmt.Errorf("enum parameter ID cannot be zero")
	}

	if value.Value == nil {
		return 0, fmt.Errorf("enum value cannot be nil")
	}

	if value.ValueType == "" {
		return 0, fmt.Errorf("enum value type cannot be empty")
	}

	id, err := s.repo.InsertEnumValue(ctx, value)
	if err != nil {
		return 0, fmt.Errorf("failed to create enum value: %w", err)
	}

	return id, nil
}

// ListEnumValues получает список значений перечисления для заданного поля и типа.
func (s *EnumValuesService) ListEnumValues(ctx context.Context, fieldName string, valueType entities.EnumType) ([]entities.EnumValue, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if fieldName == "" {
		return nil, fmt.Errorf("field name cannot be empty")
	}

	if valueType == "" {
		return nil, fmt.Errorf("value type cannot be empty")
	}

	values, err := s.repo.GetEnumValues(ctx, fieldName, valueType)
	if err != nil {
		return nil, fmt.Errorf("failed to get enum values: %w", err)
	}

	return values, nil
}

// UpdateEnumValue обновляет существующее значение перечисления по его ID.
func (s *EnumValuesService) UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return fmt.Errorf("invalid enum value ID")
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	repositoryUpdates := maps.Clone(updates)
	val, ok := repositoryUpdates["value"]
	if ok {
		delete(repositoryUpdates, "value")
		repositoryUpdates["value_raw"] = val
	}

	err := s.repo.UpdateEnumValue(ctx, id, repositoryUpdates)

	if err != nil {
		return fmt.Errorf("failed to update enum value: %w", err)
	}

	return nil
}

// DeleteEnumValue удаляет значение перечисления по его ID.
func (s *EnumValuesService) DeleteEnumValue(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return fmt.Errorf("invalid enum value ID")
	}

	err := s.repo.DeleteEnumValueByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete enum value: %w", err)
	}

	return nil
}
