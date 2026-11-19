package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ClassRepository interface {
	// GetClasses получает все классы из базы данных.
	GetClasses(ctx context.Context) ([]entities.Class, error)

	// InsertClass вставляет новую категорию в базу данных.
	InsertClass(ctx context.Context, class entities.Class) (*entities.Class, error)

	// GetClassByID получает класс по его ID.
	GetClassByID(ctx context.Context, id int64) (*entities.Class, error)

	// UpdateClass обновляет поля класса по его ID.
	UpdateClass(ctx context.Context, id int64, classUpdates entities.Class) (*entities.Class, error)

	// DeleteClass удаляет класс.
	DeleteClass(ctx context.Context, id int64) error

	// GetClassByName получает класс по его имени.
	GetClassByName(ctx context.Context, name string) (*entities.Class, error)
}

type ClassService struct {
	repo           ClassRepository
	parametersRepo ClassParametersRepository
}

// NewClassService создает сервис классов изделий и их параметров.
func NewClassService(repo ClassRepository, parametersRepo ClassParametersRepository) *ClassService {
	return &ClassService{repo: repo, parametersRepo: parametersRepo}
}

func (c *ClassService) ensureRepo() error {
	if c == nil {
		return errors.New("class service is nil")
	}
	if c.repo == nil {
		return errors.New("class repository is nil")
	}
	return nil
}

func (c *ClassService) GetAllClasses(ctx context.Context) ([]entities.Class, error) {
	if err := c.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	classes, err := c.repo.GetClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}

	return classes, nil
}

func (c *ClassService) CreateClass(ctx context.Context, class *entities.Class) (*entities.Class, error) {
	if err := c.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if class == nil {
		return nil, apperrors.BadRequest("class is nil", nil)
	}

	if class.Name == "" {
		return nil, apperrors.BadRequest("class name is required", nil)
	}

	if class.ParentID != nil && *class.ParentID == 0 {
		class.ParentID = nil
	}

	if class.ParentID != nil {
		parent, err := c.repo.GetClassByID(ctx, *class.ParentID)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && parent == nil) {
			return nil, apperrors.BadRequest("parent class not found", err)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get parent class: %w", err)
		}
	}

	class, err := c.repo.InsertClass(ctx, *class)
	if err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	return class, nil
}

func (c *ClassService) GetClassByID(ctx context.Context, id int64) (*entities.Class, error) {
	if err := c.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return nil, apperrors.BadRequest("invalid class id", nil)
	}

	class, err := c.repo.GetClassByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && class == nil) {
		return nil, apperrors.NotFound("Class not found", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}

	return class, nil
}

func (c *ClassService) UpdateClass(ctx context.Context, id int64, updates map[string]any) (*entities.Class, error) {
	if err := c.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return nil, apperrors.BadRequest("invalid class id", nil)
	}

	if len(updates) == 0 {
		return nil, apperrors.BadRequest("no fields to update", nil)
	}

	class, err := c.ValidateClassUpdates(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to validate class updates: %w", err)
	}

	class, err = c.repo.UpdateClass(ctx, id, *class)
	if err != nil {
		return nil, fmt.Errorf("failed to update class: %w", err)
	}

	return class, nil
}

func (c *ClassService) DeleteClassByID(ctx context.Context, id int64) error {
	if err := c.ensureRepo(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return apperrors.BadRequest("invalid class id", nil)
	}

	err := c.repo.DeleteClass(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}

	return nil
}

func (c *ClassService) GetParentClass(ctx context.Context, id int64) (*entities.Class, error) {
	if err := c.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	class, err := c.GetClassByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}

	if class.ParentID == nil {
		return nil, nil
	}

	parent, err := c.GetClassByID(ctx, *class.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent class: %w", err)
	}

	return parent, nil
}

func (c *ClassService) GetChildClasses(ctx context.Context, id int64) ([]entities.Class, error) {
	if err := c.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return nil, apperrors.BadRequest("invalid class id", nil)
	}

	classes, err := c.repo.GetClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}
	children := make([]entities.Class, 0)
	parentExists := false
	for _, class := range classes {
		if class.ID == id {
			parentExists = true
		}
		if class.ParentID != nil && *class.ParentID == id {
			children = append(children, class)
		}
	}
	if !parentExists {
		return nil, apperrors.NotFound("Class not found", nil)
	}
	return children, nil
}

func (c *ClassService) ensureClassParentIsNotDescendant(ctx context.Context, classID, parentID int64) error {
	visited := make(map[int64]struct{})
	currentID := parentID

	for currentID != 0 {
		if _, seen := visited[currentID]; seen {
			return errors.New("class hierarchy contains a cycle")
		}
		visited[currentID] = struct{}{}

		class, err := c.repo.GetClassByID(ctx, currentID)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && class == nil) {
			return apperrors.NotFound("Parent class not found", err)
		}
		if err != nil {
			return fmt.Errorf("failed to get parent class: %w", err)
		}

		if class.ID == classID {
			return apperrors.BadRequest("class parent creates a cycle", nil)
		}

		if class.ParentID == nil {
			break
		}
		currentID = *class.ParentID
	}

	return nil
}

func (c *ClassService) ValidateClassUpdates(ctx context.Context, id int64, updates map[string]any) (*entities.Class, error) {
	var ok bool
	class, err := c.GetClassByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}

	for key, value := range updates {
		switch key {
		case "name":
			class.Name, ok = value.(string)
			if !ok {
				return nil, apperrors.BadRequest("name must be a string", nil)
			}
			if class.Name == "" {
				return nil, apperrors.BadRequest("class name is required", nil)
			}

		case "parent_id":
			parentID, ok := value.(int64)
			if !ok || parentID <= 0 {
				return nil, apperrors.BadRequest("parent_id must be a positive int64", nil)
			}

			if err := c.ensureClassParentIsNotDescendant(ctx, id, parentID); err != nil {
				return nil, fmt.Errorf("failed to update parent class: %w", err)
			}
			class.ParentID = &parentID

		case "range_param_ids":
			rangeParamIDs, ok := value.([]int64)
			if !ok {
				return nil, apperrors.BadRequest("range_param_ids must be a slice of int64", nil)
			}
			class.RangeParamIDs = rangeParamIDs

		case "enum_param_ids":
			enumParamIDs, ok := value.([]int64)
			if !ok {
				return nil, apperrors.BadRequest("enum_param_ids must be a slice of int64", nil)
			}
			class.EnumParamIDs = enumParamIDs
		default:
			return nil, apperrors.BadRequest(fmt.Sprintf("unknown field: %s", key), nil)
		}
	}

	return class, nil
}
