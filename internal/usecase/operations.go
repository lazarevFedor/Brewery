package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/operation"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OperationRepository определяет доступ к хранилищу технологических операций.
type OperationRepository interface {
	// InsertOperation сохраняет операцию и возвращает ее с присвоенным ID.
	InsertOperation(ctx context.Context, op *entities.Operation) (*entities.Operation, error)

	// GetOperations возвращает операции с пагинацией и необязательными фильтрами по материалу и продукту.
	GetOperations(ctx context.Context, limit, offset int32, materialID, productID *int64) ([]entities.Operation, error)

	// GetOperationByID возвращает операцию по ID.
	GetOperationByID(ctx context.Context, id int64) (*entities.Operation, error)

	// UpdateOperation обновляет указанные поля операции и возвращает результат.
	UpdateOperation(ctx context.Context, id int64, updates map[string]any) (*entities.Operation, error)

	// DeleteOperation удаляет операцию по ID.
	DeleteOperation(ctx context.Context, id int64) error

	// GetProductOperations возвращает операции для указанного продукта.
	GetProductOperations(ctx context.Context, productID int64) ([]entities.Operation, error)
}

// NewOperationRepository создает репозиторий технологических операций для PostgreSQL.
func NewOperationRepository(pool *pgxpool.Pool) OperationRepository {
	return operation.NewRepository(pool)
}
