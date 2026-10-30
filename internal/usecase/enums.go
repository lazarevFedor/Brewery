// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnumRepository определяет контракт для взаимодействия с классами и значениями перечислений
type EnumRepository interface {

	// InsertEnumClass сохраняет новую сущность EnumClass в хранилище.
	InsertEnumClass(ctx context.Context, enumClass entities.EnumClass) (uint, error)

	// UpdateEnumClass обновляет сущность EnumClass в хранилище.
	UpdateEnumClass(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumClassByID удаляет сущность EnumClass из хранилища.
	DeleteEnumClassByID(ctx context.Context, id uint) error

	// GetEnumClasses получает список сущностей EnumClass по заданным имени таблицы и поля.
	GetEnumClasses(ctx context.Context, field string) ([]entities.EnumClass, error)

	// InsertEnumValue сохраняет новую сущность EnumValue в хранилище.
	InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error)

	// UpdateEnumValue обновляет сущность EnumValue в хранилище.
	UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumValueByID удаляет сущность EnumValue из хранилища.
	DeleteEnumValueByID(ctx context.Context, id uint) error

	// GetEnumValues получает список сущностей EnumValue по заданным имени поля и типу значения.
	GetEnumValues(ctx context.Context, field string, valueType entities.EnumType) ([]entities.EnumValue, error)

	// GetEnumValuesByClassID получает список сущностей EnumValue по заданному ID класса перечисления.
	GetEnumValuesByClassID(ctx context.Context, classID uint) ([]entities.EnumValue, error)

	// GetEnumClassByID получает сущность EnumClass по заданному ID класса перечисления.
	GetEnumClassByID(ctx context.Context, id uint) (*entities.EnumClass, error)
}

// NewEnumRepository создает новый экземпляр EnumRepository с переданным пулом соединений.
func NewEnumRepository(pool *pgxpool.Pool) EnumRepository {
	return &repository.EnumPostgres{Pool: pool}
}

// EnumService реализует интерфейс EnumService.
type EnumService struct {
	Repo EnumRepository
}

// CreateEnum создает новое перечисление и возвращает его ID.
func (s *EnumService) CreateEnum(ctx context.Context, enum entities.EnumClass) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	id, err := s.Repo.InsertEnumClass(ctx, enum)
	if err != nil {
		return 0, fmt.Errorf("failed to create enum class: %w", err)
	}

	return id, nil
}

// GetEnum получает список перечислений для заданного поля и типа.
func (s *EnumService) GetEnum(ctx context.Context, fieldName string) ([]entities.EnumClass, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	enums, err := s.Repo.GetEnumClasses(ctx, fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to get enums: %w", err)
	}

	return enums, nil
}

// UpdateEnum обновляет существующее перечисление по его ID.
func (s *EnumService) UpdateEnum(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	err := s.Repo.UpdateEnumClass(ctx, id, updates)

	if err != nil {
		return fmt.Errorf("failed to create enum: %w", err)
	}

	return nil
}

// DeleteEnum удаляет перечисление по его ID.
func (s *EnumService) DeleteEnum(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	err := s.Repo.DeleteEnumClassByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete enum: %w", err)
	}

	return nil
}

// CreateEnumValue создает новое значение перечисления и возвращает его ID.
func (s *EnumService) CreateEnumValue(ctx context.Context, enum entities.EnumValue) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	id, err := s.Repo.InsertEnumValue(ctx, enum)
	if err != nil {
		return 0, fmt.Errorf("failed to create enum value: %w", err)
	}

	return id, nil
}

// GetEnumValue получает список значений перечисления для заданного поля и типа.
func (s *EnumService) GetEnumValue(ctx context.Context, fieldName string, valueType entities.EnumType) ([]entities.EnumValue, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	enums, err := s.Repo.GetEnumValues(ctx, fieldName, valueType)
	if err != nil {
		return nil, fmt.Errorf("failed to get enum values: %w", err)
	}

	return enums, nil
}

// UpdateEnumValue обновляет существующее значение перечисления по его ID.
func (s *EnumService) UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	val, ok := updates["value"]
	if ok {
		delete(updates, "value")
		updates["value_raw"] = val
	}

	err := s.Repo.UpdateEnumValue(ctx, id, updates)

	if err != nil {
		return fmt.Errorf("failed to create enum value: %w", err)
	}

	return nil
}

// DeleteEnumValue удаляет значение перечисления по его ID.
func (s *EnumService) DeleteEnumValue(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	err := s.Repo.DeleteEnumValueByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete enum value: %w", err)
	}

	return nil
}
