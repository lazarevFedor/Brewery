package usecase

import (
	"Brewery/internal/entities"
	"context"

	"github.com/jackc/pgx/v5"
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