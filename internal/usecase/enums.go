package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"fmt"
)

type EnumService interface {
	CreateEnum(ctx context.Context, enum entities.EnumClass) (uint, error)
	GetEnum(ctx context.Context, entityName, fieldName string) ([]entities.EnumClass, error)
	UpdateEnum(ctx context.Context, id uint, updates map[string]any) error
	DeleteEnum(ctx context.Context, id uint) error

	CreateEnumValue(ctx context.Context, enum entities.EnumValue) (uint, error)
	GetEnumValue(ctx context.Context, entity, field string, valueType entities.EnumType) ([]entities.EnumValue, error)
	UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error
	DeleteEnumValue(ctx context.Context, id uint) error
}

type enumService struct {
	repo repository.EnumRepository
}

func NewEnumService(enumRepo repository.EnumRepository) EnumService {
	return &enumService{
		repo: enumRepo,
	}
}

func (s *enumService) CreateEnum(ctx context.Context, enum entities.EnumClass) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	id, err := s.repo.InsertEnumClass(ctx, enum)
	if err != nil {
		return 0, fmt.Errorf("failed to create enum class: %w", err)
	}

	return id, nil
}

func (s *enumService) GetEnum(ctx context.Context, entityName, fieldName string) ([]entities.EnumClass, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	enums, err := s.repo.GetEnumClasses(ctx, entityName, fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to get enums: %w", err)
	}

	return enums, nil
}

func (s *enumService) UpdateEnum(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	err := s.repo.UpdateEnumClass(ctx, id, updates)

	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

func (s *enumService) DeleteEnum(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	err := s.repo.DeleteEnumClassByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	return nil
}

func (s *enumService) CreateEnumValue(ctx context.Context, enum entities.EnumValue) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	id, err := s.repo.InsertEnumValue(ctx, enum)
	if err != nil {
		return 0, fmt.Errorf("failed to create enum value: %w", err)
	}

	return id, nil
}

func (s *enumService) GetEnumValue(ctx context.Context, entityName, fieldName string, valueType entities.EnumType) ([]entities.EnumValue, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	enums, err := s.repo.GetEnumValues(ctx, entityName, fieldName, valueType)
	if err != nil {
		return nil, fmt.Errorf("failed to get enum values: %w", err)
	}

	return enums, nil
}

func (s *enumService) UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	val, ok := updates["value"]
	if ok {
		delete(updates, "value")
		updates["value_raw"] = val
	}

	err := s.repo.UpdateEnumValue(ctx, id, updates)

	if err != nil {
		return fmt.Errorf("failed to create enum value: %w", err)
	}

	return nil
}

func (s *enumService) DeleteEnumValue(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	err := s.repo.DeleteEnumValueByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete enum value: %w", err)
	}

	return nil
}
