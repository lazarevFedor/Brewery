// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"fmt"
	"math"

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

	if value.EnumClassID <= 0 {
		return 0, apperrors.BadRequest("invalid enum parameter ID", nil)
	}

	if value.Value == nil {
		return 0, apperrors.BadRequest("enum value cannot be nil", nil)
	}

	if !value.ValueType.Valid() {
		return 0, apperrors.BadRequest("enum_type must be string, int, or float", nil)
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
		return nil, apperrors.BadRequest("field name cannot be empty", nil)
	}

	if !valueType.Valid() {
		return nil, apperrors.BadRequest("enum_type must be string, int, or float", nil)
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
		return apperrors.BadRequest("invalid enum value ID", nil)
	}

	if len(updates) == 0 {
		return apperrors.BadRequest("no updates provided", nil)
	}

	repositoryUpdates := make(map[string]any, len(updates))
	for field, value := range updates {
		switch field {
		case "value":
			if _, isString := value.(string); !isString {
				number, isNumber := toFloat64Value(value)
				if !isNumber || math.IsNaN(number) || math.IsInf(number, 0) {
					return apperrors.BadRequest("value must be a string or finite number", nil)
				}
			}
			repositoryUpdates["value_raw"] = value
		case "position":
			position, ok := toFloat64Value(value)
			if !ok || position != math.Trunc(position) || position < math.MinInt32 || position > math.MaxInt32 {
				return apperrors.BadRequest("position must be a 32-bit integer", nil)
			}
			repositoryUpdates["position"] = int32(position)
		default:
			return apperrors.BadRequest(fmt.Sprintf("unknown enum value update field %q", field), nil)
		}
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
		return apperrors.BadRequest("invalid enum value ID", nil)
	}

	err := s.repo.DeleteEnumValueByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete enum value: %w", err)
	}

	return nil
}
