package usecase

import (
	"Brewery/internal/entities"
	"context"
)

type ClassRepository interface {
	// GetClasses получает все классы из базы данных.
	GetClasses(ctx context.Context) ([]entities.Class, error)

	// InsertClass вставляет новую категорию в базу данных.
	InsertClass(ctx context.Context, class entities.Class) (*entities.Class, error)

	// GetClassByID получает класс по его ID.
	GetClassByID(ctx context.Context, id int64) (*entities.Class, error)

	// UpdateClassName обновляет имя класса по его ID.
	UpdateClassName(ctx context.Context, id int64, name string) (*entities.Class, error)

	// UpdateClassParent обновляет родительский класс ID.
	UpdateClassParent(ctx context.Context, id int64, parentID int64) (*entities.Class, error)

	// DeleteClassByID удаляет класс по его ID.
	DeleteClass(ctx context.Context, id int64) error

	// GetClassByName получает класс по его имени.
	GetClassByName(ctx context.Context, name string) (*entities.Class, error)
}
