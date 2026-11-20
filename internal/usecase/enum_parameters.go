// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnumParametersRepository определяет операции хранения перечисляемых параметров.
type EnumParametersRepository interface {
	// InsertEnumParameter сохраняет новый перечисляемый параметр в хранилище.
	InsertEnumParameter(ctx context.Context, parameter entities.EnumClass) (uint, error)

	// UpdateEnumParameter обновляет перечисляемый параметр в хранилище.
	UpdateEnumParameter(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumParameterByID удаляет перечисляемый параметр из хранилища.
	DeleteEnumParameterByID(ctx context.Context, id uint) error

	// GetEnumParameters получает перечисляемые параметры по имени поля.
	GetEnumParameters(ctx context.Context, field string) ([]entities.EnumClass, error)

	// GetEnumParameterByID получает перечисляемый параметр по ID.
	GetEnumParameterByID(ctx context.Context, id uint) (*entities.EnumClass, error)
}

// NewEnumParametersRepository создает репозиторий перечисляемых параметров.
func NewEnumParametersRepository(pool *pgxpool.Pool) EnumParametersRepository {
	return repository.NewEnumParametersPostgres(pool)
}

// EnumParametersService реализует бизнес-логику перечисляемых параметров.
type EnumParametersService struct {
	repo EnumParametersRepository
}

// NewEnumParametersService создает сервис перечисляемых параметров.
func NewEnumParametersService(repo EnumParametersRepository) *EnumParametersService {
	return &EnumParametersService{repo: repo}
}

// CreateEnumParameter создает новый перечисляемый параметр и возвращает его ID.
func (s *EnumParametersService) CreateEnumParameter(ctx context.Context, parameter entities.EnumClass) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}
	if !entities.EnumType(parameter.Type).Valid() {
		return 0, apperrors.BadRequest("enum_type must be string, int, or float", nil)
	}

	id, err := s.repo.InsertEnumParameter(ctx, parameter)
	if err != nil {
		return 0, fmt.Errorf("failed to create enum parameter: %w", err)
	}

	return id, nil
}

// ListEnumParameters получает перечисляемые параметры для заданного поля.
func (s *EnumParametersService) ListEnumParameters(ctx context.Context, fieldName string) ([]entities.EnumClass, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if fieldName == "" {
		return nil, apperrors.BadRequest("field name cannot be empty", nil)
	}

	parameters, err := s.repo.GetEnumParameters(ctx, fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to get enum parameters: %w", err)
	}

	return parameters, nil
}

// UpdateEnumParameter обновляет перечисляемый параметр по его ID.
func (s *EnumParametersService) UpdateEnumParameter(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return apperrors.BadRequest("invalid enum parameter ID", nil)
	}

	if len(updates) == 0 {
		return apperrors.BadRequest("no updates provided", nil)
	}

	for field, value := range updates {
		switch field {
		case "field_name":
			if name, ok := value.(string); !ok || name == "" {
				return apperrors.BadRequest("field_name must be a non-empty string", nil)
			}
		case "unit":
			if _, ok := value.(string); !ok && value != nil {
				return apperrors.BadRequest("unit must be a string or null", nil)
			}
		case "inheritable":
			if _, ok := value.(bool); !ok {
				return apperrors.BadRequest("inheritable must be a boolean", nil)
			}
		case "enum_type":
			return apperrors.BadRequest("enum_type cannot be changed after creation", nil)
		default:
			return apperrors.BadRequest(fmt.Sprintf("unknown enum parameter update field %q", field), nil)
		}
	}

	err := s.repo.UpdateEnumParameter(ctx, id, updates)

	if err != nil {
		return fmt.Errorf("failed to update enum parameter: %w", err)
	}

	return nil
}

// DeleteEnumParameter удаляет перечисляемый параметр по его ID.
func (s *EnumParametersService) DeleteEnumParameter(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return apperrors.BadRequest("invalid enum parameter ID", nil)
	}

	err := s.repo.DeleteEnumParameterByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete enum parameter: %w", err)
	}

	return nil
}
