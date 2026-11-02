package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClassRepository interface {
	// GetClasses получает все классы из базы данных.
	GetClasses(ctx context.Context) ([]entities.Class, error)

	// InsertClass вставляет новую категорию в базу данных.
	InsertClass(ctx context.Context, tx pgx.Tx, class entities.Class) (uint, error)

	// GetClassByID получает класс по его ID.
	GetClassByID(ctx context.Context, id uint) (*entities.Class, error)

	// UpdateClass обновляет поля класса по его ID.
	UpdateClass(ctx context.Context, id uint, updates map[string]any) error

	// DeleteClassByID удаляет класс по его ID.
	DeleteClassByID(ctx context.Context, id uint) error

	// GetClassID получает ID класса по его имени.
	GetClassID(ctx context.Context, tx pgx.Tx, ctgName string) (uint, error)
}

// NewClassRepository создает новый экземпляр ClassRepository с переданным пулом соединений.
func NewClassRepository(pool *pgxpool.Pool) ClassRepository {
	return &repository.ClassPostgres{Pool: pool}
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

func (c *ClassService) CreateClass(ctx context.Context, class *entities.Class) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if class == nil {
		return 0, errors.New("class is nil")
	}

	if class.Name == "" {
		return 0, errors.New("class name is required")
	}

	if class.ParentID == 0 {
		classes, err := c.repo.GetClasses(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to check existing classes: %w", err)
		}
		for _, c := range classes {
			if c.ParentID == 0 {
				return 0, errors.New("root class already exists")
			}
		}
	} else {
		parent, err := c.repo.GetClassByID(ctx, uint(class.ParentID))
		if err != nil {
			return 0, fmt.Errorf("failed to get parent class: %w", err)
		}
		if parent == nil {
			return 0, errors.New("parent class not found")
		}
	}

	id, err := c.repo.InsertClass(ctx, nil, *class)
	if err != nil {
		return 0, fmt.Errorf("failed to create class: %w", err)
	}

	return id, nil
}

func (c *ClassService) GetClassByID(ctx context.Context, id uint) (*entities.Class, error) {
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

func (c *ClassService) UpdateClass(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid class id")
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	parentID, ok := updates["parent_id"]
	if ok {
		parentIDFloat, ok := parentID.(float64)
		if ok {
			if parentIDFloat == 0 {
				updates["parent_id"] = nil
			} else {
				err := c.ensureClassParentIsNotDescendant(ctx, id, uint(parentIDFloat))
				if err != nil {
					return fmt.Errorf("failed to update parent class: %w", err)
				}
			}
		}
	}

	err := c.repo.UpdateClass(ctx, id, updates)
	if err != nil {
		return fmt.Errorf("failed to update class: %w", err)
	}

	return nil
}

func (c *ClassService) DeleteClassByID(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid class id")
	}

	err := c.repo.DeleteClassByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}

	return nil
}

func (c *ClassService) GetParentClass(ctx context.Context, id uint) (*entities.Class, error) {
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

	parent, err := c.repo.GetClassByID(ctx, uint(class.ParentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get parent class: %w", err)
	}

	return parent, nil
}

func (c *ClassService) GetChildClasses(ctx context.Context, id uint) ([]entities.Class, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	classes, err := c.repo.GetClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}
	children := make([]entities.Class, 0)
	for _, c := range classes {
		if c.ParentID != 0 && uint(c.ParentID) == id {
			children = append(children, c)
		}
	}
	return children, nil
}

func (c *ClassService) ensureClassParentIsNotDescendant(ctx context.Context, classID, parentID uint) error {
	visited := make(map[uint]struct{})
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

		if class.ID == int(classID) {
			return errors.New("class parent creates a cycle")
		}

		currentID = uint(class.ParentID)
	}

	return nil
}
