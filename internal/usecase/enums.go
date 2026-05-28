// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"
)

// EnumService определяет интерфейс для работы с перечислениями и их значениями.
type EnumService interface {

	// CreateEnum создает новое перечисление и возвращает его ID.
	CreateEnum(ctx context.Context, enum entities.EnumClass) (uint, error)

	// GetEnum получает список перечислений для заданного поля и типа.
	GetEnum(ctx context.Context, entityName, fieldName string) ([]entities.EnumClass, error)

	// UpdateEnum обновляет существующее перечисление по его ID.
	UpdateEnum(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnum удаляет перечисление по его ID.
	DeleteEnum(ctx context.Context, id uint) error

	// CreateEnumValue создает новое значение перечисления и возвращает его ID.
	CreateEnumValue(ctx context.Context, enum entities.EnumValue) (uint, error)

	// GetEnumValue получает список значений перечисления для заданного поля и типа.
	GetEnumValue(ctx context.Context, entity, field string, valueType entities.EnumType) ([]entities.EnumValue, error)

	// UpdateEnumValue обновляет существующее значение перечисления по его ID.
	UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumValue удаляет значение перечисления по его ID.
	DeleteEnumValue(ctx context.Context, id uint) error
}

// enumService реализует интерфейс EnumService.
type enumService struct {
	repo repository.EnumRepository
}

// NewEnumService создает новый EnumService с заданным репозиторием.
func NewEnumService(enumRepo repository.EnumRepository) EnumService {
	return &enumService{
		repo: enumRepo,
	}
}

// CreateEnum создает новое перечисление и возвращает его ID.
func (s *enumService) CreateEnum(ctx context.Context, enum entities.EnumClass) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	id, err := s.repo.InsertEnumClass(ctx, enum)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetEnum получает список перечислений для заданного поля и типа.
func (s *enumService) GetEnum(ctx context.Context, entityName, fieldName string) ([]entities.EnumClass, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	enums, err := s.repo.GetEnumClasses(ctx, entityName, fieldName)
	if err != nil {
		return nil, err
	}

	return enums, nil
}

// UpdateEnum обновляет существующее перечисление по его ID.
func (s *enumService) UpdateEnum(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	if len(updates) == 0 {
		return apperrors.BadRequest("no updates provided", errors.New("no updates provided"))
	}

	err := s.repo.UpdateEnumClass(ctx, id, updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteEnum удаляет перечисление по его ID.
func (s *enumService) DeleteEnum(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	err := s.repo.DeleteEnumClassByID(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// CreateEnumValue создает новое значение перечисления и возвращает его ID.
func (s *enumService) CreateEnumValue(ctx context.Context, enum entities.EnumValue) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	id, err := s.repo.InsertEnumValue(ctx, enum)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetEnumValue получает список значений перечисления для заданного поля и типа.
func (s *enumService) GetEnumValue(ctx context.Context, entityName, fieldName string, valueType entities.EnumType) ([]entities.EnumValue, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	enums, err := s.repo.GetEnumValues(ctx, entityName, fieldName, valueType)
	if err != nil {
		return nil, err
	}

	return enums, nil
}

// UpdateEnumValue обновляет существующее значение перечисления по его ID.
func (s *enumService) UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	val, ok := updates["value"]
	if ok {
		delete(updates, "value")
		updates["value_raw"] = val
	}

	err := s.repo.UpdateEnumValue(ctx, id, updates)

	if err != nil {
		return err
	}

	return nil
}

// DeleteEnumValue удаляет значение перечисления по его ID.
func (s *enumService) DeleteEnumValue(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	err := s.repo.DeleteEnumValueByID(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
