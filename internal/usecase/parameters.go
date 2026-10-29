package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"
)

// ParametersService - сервис для управления параметрами.
type ParametersService interface {

	// CreateRange создает числовой параметр.
	CreateRange(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error)

	// UpdateRange обновляет числовой параметр. Единственные поля, которые можно обновлять - min_val, max_val и inheritable.
	UpdateRange(ctx context.Context, id uint, updates map[string]any) (*entities.RangeParameter, error)

	// DeleteRange удаляет числовой параметр.
	DeleteRange(ctx context.Context, id uint) (*entities.RangeParameter, error)

	// CreateEnum создает параметр типа перечисление.
	CreateEnum(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error)

	// UpdateEnum обновляет перечисление. Единственное поле, которое можно обновлять - inheritable.
	UpdateEnum(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error)

	// DeleteEnum удаляет перечисление. Если параметр был применён к категориям, он будет удалён из них, но категории сохранятся.
	DeleteEnum(ctx context.Context, id uint) (*entities.EnumParameter, error)

	// ListParameters возвращает все параметры, применимые к категории.
	// categoryID = 0 - без фильтра по категории, paramType = "" - все типы, "numeric" - только числовые, "enum" - только перечисления
	ListParameters(ctx context.Context, categoryID uint, parameterType int) ([]entities.RangeParameter, []entities.EnumParameter, error)

	// ApplyToCategory применяет указанные параметры к категории. Возвращает количество добавленных параметров.
	ApplyToCategory(ctx context.Context, categoryID uint, numericIDs, enumIDs []int) (int, error)
}

// parametersService - реализация ParametersService.
type parametersService struct {
	paramRepo repository.ParameterRepository
	enumRepo  repository.EnumRepository
}

// NewParametersService создает новый экземпляр ParametersService.
func NewParametersService(paramRepo repository.ParameterRepository, enumRepo repository.EnumRepository) ParametersService {
	return &parametersService{
		paramRepo: paramRepo,
		enumRepo:  enumRepo,
	}
}

// allowedParametersToUpdate содержит список полей, которые можно обновлять в числовом параметре.
var allowedParametersToUpdate = map[string]struct{}{
	"min_val":     {},
	"max_val":     {},
	"inheritable": {},
}

// CreateRange создает числовой параметр.
func (s *parametersService) CreateRange(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
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

	created, err := s.paramRepo.InsertNumericParameter(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create numeric parameter: %w", err)
	}

	return created, nil
}

// UpdateRange обновляет числовой параметр. Единственные поля, которые можно обновлять - min_val, max_val и inheritable.
func (s *parametersService) UpdateRange(ctx context.Context, id uint, updates map[string]any) (*entities.RangeParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	for k := range updates {
		if _, ok := allowedParametersToUpdate[k]; !ok {
			return nil, fmt.Errorf("unknown parameter: %s", k)
		}
	}

	updated, err := s.paramRepo.UpdateNumericParameter(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update numeric parameter: %w", err)
	}

	return updated, nil
}

// DeleteRange удаляет числовой параметр.
func (s *parametersService) DeleteRange(ctx context.Context, id uint) (*entities.RangeParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	deleted, err := s.paramRepo.DeleteNumericParameter(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete numeric parameter: %w", err)
	}

	return deleted, nil
}

// CreateEnum создает параметр типа перечисление.
func (s *parametersService) CreateEnum(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	enumClass, err := s.enumRepo.GetEnumClassByID(ctx, param.EnumClassID)
	if err != nil || enumClass == nil {
		return nil, fmt.Errorf("failed to find enum_class: %w", err)
	}

	created, err := s.paramRepo.InsertEnumParameter(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create enum parameter: %w", err)
	}

	return created, nil
}

// UpdateEnum обновляет перечисление. Единственное поле, которое можно обновлять - inheritable.
func (s *parametersService) UpdateEnum(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if _, ok := updates["inheritable"]; len(updates) > 1 || !ok {
		return nil, errors.New("inheritable is required")
	}

	updated, err := s.paramRepo.UpdateEnumParameter(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update enum parameter: %w", err)
	}

	return updated, nil
}

// DeleteEnum удаляет перечисление. Если параметр был применён к категориям, он будет удалён из них, но категории сохранятся.
func (s *parametersService) DeleteEnum(ctx context.Context, id uint) (*entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	deleted, err := s.paramRepo.DeleteEnumParameter(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete enum parameter: %w", err)
	}

	return deleted, nil
}

// ListParameters возвращает все параметры, применимые к категории.
func (s *parametersService) ListParameters(ctx context.Context, categoryID uint, parameterType int) ([]entities.RangeParameter, []entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("request cancelled: %w", err)
	}

	numeric, enum, err := s.paramRepo.GetParameters(ctx, categoryID, parameterType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get parameters: %w", err)
	}

	return numeric, enum, nil
}

// ApplyToCategory применяет указанные параметры к категории. Возвращает количество добавленных параметров.
func (s *parametersService) ApplyToCategory(ctx context.Context, categoryID uint, numericIDs, enumIDs []int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	added, err := s.paramRepo.ApplyParameters(ctx, categoryID, numericIDs, enumIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to apply parameters: %w", err)
	}

	return added, nil
}
