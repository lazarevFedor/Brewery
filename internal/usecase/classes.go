package usecase

import (
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"
)

type ClassRepository interface {
	// GetClasses получает все классы из базы данных.
	GetClasses(ctx context.Context) ([]entities.Class, error)

	// InsertClass вставляет новую категорию в базу данных.
	InsertClass(ctx context.Context, class entities.Class) (*entities.Class, error)

	// GetClassByID получает класс по его ID.
	GetClassByID(ctx context.Context, id int64) (*entities.Class, error)

	// UpdateClass обновляет поля класса по его ID.
	UpdateClass(ctx context.Context, classUpdates entities.Class) (*entities.Class, error)

	// DeleteClass удаляет класс.
	DeleteClass(ctx context.Context, id int64) error

	// GetClassByName получает класс по его имени.
	GetClassByName(ctx context.Context, name string) (*entities.Class, error)
}

type ClassService struct {
	repo ClassRepository
}

func (c *ClassService) GetAllClasses(ctx context.Context) ([]entities.Class, error) {
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
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if class == nil {
		return nil, errors.New("class is nil")
	}

	if class.Name == "" {
		return nil, errors.New("class name is required")
	}

	if class.ParentID == 0 {
		classes, err := c.repo.GetClasses(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing classes: %w", err)
		}
		for _, c := range classes {
			if c.ParentID == 0 {
				return nil, errors.New("root class already exists")
			}
		}
	} else {
		parent, err := c.repo.GetClassByID(ctx, class.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get parent class: %w", err)
		}
		if parent == nil {
			return nil, errors.New("parent class not found")
		}
	}

	class, err := c.repo.InsertClass(ctx, *class)
	if err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	return class, nil
}

func (c *ClassService) GetClassByID(ctx context.Context, id int64) (*entities.Class, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, errors.New("invalid class id")
	}

	class, err := c.repo.GetClassByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}

	return class, nil
}

func (c *ClassService) UpdateClass(ctx context.Context, id int64, updates map[string]any) (*entities.Class, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, errors.New("invalid class id")
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	class, err := c.ValidateClassUpdates(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to validate class updates: %w", err)
	}

	class, err = c.repo.UpdateClass(ctx, *class)
	if err != nil {
		return nil, fmt.Errorf("failed to update class: %w", err)
	}

	return class, nil
}

func (c *ClassService) DeleteClassByID(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid class id")
	}

	err := c.repo.DeleteClass(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}

	return nil
}

func (c *ClassService) GetParentClass(ctx context.Context, id int64) (*entities.Class, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	class, err := c.repo.GetClassByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}

	if class.ParentID == 0 {
		return nil, nil
	}

	parent, err := c.repo.GetClassByID(ctx, class.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent class: %w", err)
	}

	return parent, nil
}

func (c *ClassService) GetChildClasses(ctx context.Context, id int64) ([]entities.Class, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	classes, err := c.repo.GetClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}
	children := make([]entities.Class, 0)
	for _, c := range classes {
		if c.ParentID != 0 && c.ParentID == id {
			children = append(children, c)
		}
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
		if err != nil {
			return fmt.Errorf("failed to get parent class: %w", err)
		}
		if class == nil {
			return errors.New("parent class not found")
		}

		if class.ID == classID {
			return errors.New("class parent creates a cycle")
		}

		currentID = class.ParentID
	}

	return nil
}

func (c *ClassService) ValidateClassUpdates(ctx context.Context, id int64, updates map[string]any) (*entities.Class, error) {
	class := &entities.Class{}
	var ok bool

	for key, value := range updates {
		if key == "name" {
			class.Name, ok = value.(string)
			if !ok {
				return nil, errors.New("name must be a string")
			}
		}

		if key == "parent_id" {
			class.ParentID, ok = value.(int64)
			if !ok {
				return nil, errors.New("parent_id must be an int64")
			}
			if class.ParentID == 0 {
				updates["parent_id"] = nil
			} else {
				err := c.ensureClassParentIsNotDescendant(ctx, id, class.ParentID)
				if err != nil {
					return nil, fmt.Errorf("failed to update parent class: %w", err)
				}
			}
		}

		if key == "range_param_ids" {
			rangeParamIDs, ok := value.([]int64)
			if !ok {
				return nil, errors.New("range_param_ids must be a slice of int64")
			}
			class.RangeParamIDs = rangeParamIDs
		}

		if key == "enum_param_ids" {
			enumParamIDs, ok := value.([]int64)
			if !ok {
				return nil, errors.New("enum_param_ids must be a slice of int64")
			}
			class.EnumParamIDs = enumParamIDs
		}
	}

	return class, nil
}
