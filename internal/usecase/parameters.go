package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"
)

type ParametersService interface {
	CreateNumeric(ctx context.Context, param *entities.NumericParameter) (*entities.NumericParameter, error)
	UpdateNumeric(ctx context.Context, id uint, updates map[string]any) (*entities.NumericParameter, error)
	DeleteNumeric(ctx context.Context, id uint) (*entities.NumericParameter, error)

	CreateEnum(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error)
	UpdateEnum(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error)
	DeleteEnum(ctx context.Context, id uint) (*entities.EnumParameter, error)

	ListParameters(ctx context.Context, categoryID uint) ([]entities.NumericParameter, []entities.EnumParameter, error)
	ApplyToCategory(ctx context.Context, categoryID uint, numericIDs, enumIDs []int) (int, error)
}

type parametersService struct {
	paramRepo repository.ParameterRepository
}

func NewParametersService(paramRepo repository.ParameterRepository) ParametersService {
	return &parametersService{
		paramRepo: paramRepo,
	}
}

func (s *parametersService) CreateNumeric(ctx context.Context, param *entities.NumericParameter) (*entities.NumericParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if param.FieldName == "" {
		return nil, errors.New("field_name is required")
	}
	if param.EntityName == "" {
		return nil, errors.New("entity_name is required")
	}
	if param.MinValue > param.MaxValue {
		return nil, errors.New("min_val must be less than or equal to max_val")
	}

	created, err := s.paramRepo.InsertNumericParameter(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create numeric parameter: %w", err)
	}

	return created, nil
}

func (s *parametersService) UpdateNumeric(ctx context.Context, id uint, updates map[string]any) (*entities.NumericParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	updated, err := s.paramRepo.UpdateNumericParameter(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update numeric parameter: %w", err)
	}

	return updated, nil
}

func (s *parametersService) DeleteNumeric(ctx context.Context, id uint) (*entities.NumericParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	deleted, err := s.paramRepo.DeleteNumericParameter(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete numeric parameter: %w", err)
	}

	return deleted, nil
}

func (s *parametersService) CreateEnum(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if param.EnumClassID == 0 {
		return nil, errors.New("enum_class_id is required")
	}

	created, err := s.paramRepo.InsertEnumParameter(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create enum parameter: %w", err)
	}

	return created, nil
}

func (s *parametersService) UpdateEnum(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	updated, err := s.paramRepo.UpdateEnumParameter(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update enum parameter: %w", err)
	}

	return updated, nil
}

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

func (s *parametersService) ListParameters(ctx context.Context, categoryID uint) ([]entities.NumericParameter, []entities.EnumParameter, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("request cancelled: %w", err)
	}

	numeric, enum, err := s.paramRepo.GetParameters(ctx, categoryID, entities.MissingType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get parameters: %w", err)
	}

	return numeric, enum, nil
}

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
