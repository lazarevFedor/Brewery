package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"
)

// ParametersService - сервис для управления параметрами.
type ParametersService interface {

	// CreateNumeric создает числовой параметр.
	CreateNumeric(ctx context.Context, param *entities.NumericParameter) (*entities.NumericParameter, error)

	// UpdateNumeric обновляет числовой параметр. Единственные поля, которые можно обновлять - min_val, max_val и inheritable.
	UpdateNumeric(ctx context.Context, id uint, updates map[string]any) (*entities.NumericParameter, error)

	// DeleteNumeric удаляет числовой параметр.
	DeleteNumeric(ctx context.Context, id uint) (*entities.NumericParameter, error)

	// CreateEnum создает параметр типа перечисление.
	CreateEnum(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error)

	// UpdateEnum обновляет перечисление. Единственное поле, которое можно обновлять - inheritable.
	UpdateEnum(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error)

	// DeleteEnum удаляет перечисление. Если параметр был применён к категориям, он будет удалён из них, но категории сохранятся.
	DeleteEnum(ctx context.Context, id uint) (*entities.EnumParameter, error)

	// ListParameters возвращает все параметры, применимые к категории.
	// categoryID = 0 - без фильтра по категории, paramType = "" - все типы, "numeric" - только числовые, "enum" - только перечисления
	ListParameters(ctx context.Context, categoryID uint, parameterType int) ([]entities.NumericParameter, []entities.EnumParameter, error)

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

// CreateNumeric создает числовой параметр.
func (s *parametersService) CreateNumeric(ctx context.Context, param *entities.NumericParameter) (*entities.NumericParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	if param.FieldName == "" {
		return nil, apperrors.BadRequest("field_name is required", errors.New("field_name is required"))
	}
	if param.EntityName == "" {
		return nil, apperrors.BadRequest("entity_name is required", errors.New("entity_name is required"))
	}
	if param.MinValue > param.MaxValue {
		return nil, apperrors.BadRequest("min_val cannot be greater than max_val", errors.New("min_val cannot be greater than max_val"))
	}
	if param.MinValue < 0 {
		return nil, apperrors.BadRequest("min_val must be non-negative", errors.New("min_val must be non-negative"))
	}

	created, err := s.paramRepo.InsertNumericParameter(ctx, param)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateNumeric обновляет числовой параметр. Единственные поля, которые можно обновлять - min_val, max_val и inheritable.
func (s *parametersService) UpdateNumeric(ctx context.Context, id uint, updates map[string]any) (*entities.NumericParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	if len(updates) == 0 {
		return nil, apperrors.BadRequest("no updates provided", errors.New("no updates provided"))
	}

	for k := range updates {
		if _, ok := allowedParametersToUpdate[k]; !ok {
			return nil, apperrors.BadRequest(fmt.Sprintf("field %s cannot be updated", k), fmt.Errorf("field %s cannot be updated", k))
		}
	}

	updated, err := s.paramRepo.UpdateNumericParameter(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// DeleteNumeric удаляет числовой параметр.
func (s *parametersService) DeleteNumeric(ctx context.Context, id uint) (*entities.NumericParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	deleted, err := s.paramRepo.DeleteNumericParameter(ctx, id)
	if err != nil {
		return nil, err
	}

	return deleted, nil
}

// CreateEnum создает параметр типа перечисление.
func (s *parametersService) CreateEnum(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	enumClass, err := s.enumRepo.GetEnumClassByID(ctx, param.EnumClassID)
	if err != nil || enumClass == nil {
		return nil, fmt.Errorf("failed to find enum_class: %w", err)
	}

	created, err := s.paramRepo.InsertEnumParameter(ctx, param)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateEnum обновляет перечисление. Единственное поле, которое можно обновлять - inheritable.
func (s *parametersService) UpdateEnum(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	if len(updates) == 0 {
		return nil, apperrors.BadRequest("no updates provided", errors.New("no updates provided"))
	}

	if _, ok := updates["inheritable"]; len(updates) > 1 || !ok {
		return nil, apperrors.BadRequest("only inheritable field can be updated", errors.New("only inheritable field can be updated"))
	}

	updated, err := s.paramRepo.UpdateEnumParameter(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// DeleteEnum удаляет перечисление. Если параметр был применён к категориям, он будет удалён из них, но категории сохранятся.
func (s *parametersService) DeleteEnum(ctx context.Context, id uint) (*entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	deleted, err := s.paramRepo.DeleteEnumParameter(ctx, id)
	if err != nil {
		return nil, err
	}

	return deleted, nil
}

// ListParameters возвращает все параметры, применимые к категории.
func (s *parametersService) ListParameters(ctx context.Context, categoryID uint, parameterType int) ([]entities.NumericParameter, []entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	numeric, enum, err := s.paramRepo.GetParameters(ctx, categoryID, parameterType)
	if err != nil {
		return nil, nil, err
	}

	return numeric, enum, nil
}

// ApplyToCategory применяет указанные параметры к категории. Возвращает количество добавленных параметров.
func (s *parametersService) ApplyToCategory(ctx context.Context, categoryID uint, numericIDs, enumIDs []int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	added, err := s.paramRepo.ApplyParameters(ctx, categoryID, numericIDs, enumIDs)
	if err != nil {
		return 0, err
	}

	return added, nil
}
