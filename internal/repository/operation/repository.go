// Package operation содержит репозиторий технологических операций.
package operation

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const operationsTable = "operations"

// operationColumns - столбцы, которые читаются из таблицы operations, в порядке scanOperation.
var operationColumns = []string{
	"id",
	"name",
	"description",
	"operation_class_id",
	"profession_id",
	"labor_cost",
	"step_number",
	"equipment_name",
	"work_center_name",
	"input_product_id",
	"output_product_id",
}

// updatableColumns - поля operations, которые можно менять через UpdateOperation.
// Ключ - ключ в map обновлений, значение - функция приведения значения к типу столбца.
var updatableColumns = map[string]func(any) (any, error){
	"name":               toText,
	"description":        toText,
	"operation_class_id": toInt64,
	"profession_id":      toInt64,
	"labor_cost":         toFloat64,
	"step_number":        toInt64,
	"equipment_name":     toText,
	"work_center_name":   toText,
}

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// Repository реализует usecase.OperationRepository для PostgreSQL.
type Repository struct {
	Pool *pgxpool.Pool
}

// NewRepository создает репозиторий технологических операций.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{Pool: pool}
}

// InsertOperation сохраняет операцию и ее параметры в одной транзакции.
// Возвращает переданную операцию с присвоенным ID.
func (repo *Repository) InsertOperation(ctx context.Context, op *entities.Operation) (*entities.Operation, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	if op == nil {
		return nil, errors.New("operation is nil")
	}

	tx, err := repo.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("InsertOperation: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // после Commit откат ничего не делает

	sql, args, err := psql.Insert(operationsTable).
		Columns(
			"name",
			"description",
			"operation_class_id",
			"profession_id",
			"labor_cost",
			"step_number",
			"equipment_name",
			"work_center_name",
			"input_product_id",
			"output_product_id",
		).
		Values(
			op.Name,
			nullableText(op.Description),
			op.ClassID,
			op.ProfessionID,
			op.LaborCost,
			op.StepNumber,
			op.Equipment.Name,
			op.Equipment.WorkCenter,
			op.Materials.MaterialID,
			op.Materials.ProductID,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("InsertOperation: build query: %w", err)
	}

	var id int64
	if err = tx.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return nil, wrapDBError("InsertOperation", err)
	}

	if err = insertParameters(ctx, tx, id, op.Parameters); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("InsertOperation: commit: %w", err)
	}

	result := *op
	result.ID = id
	return &result, nil
}

// GetOperations возвращает операции с пагинацией.
// materialID фильтрует по входному изделию, productID - по выходному; nil - без фильтра.
func (repo *Repository) GetOperations(
	ctx context.Context,
	limit, offset int32,
	materialID, productID *int64,
) ([]entities.Operation, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	builder := psql.Select(operationColumns...).
		From(operationsTable).
		OrderBy("output_product_id", "step_number", "id")
	if materialID != nil {
		builder = builder.Where(squirrel.Eq{"input_product_id": *materialID})
	}
	if productID != nil {
		builder = builder.Where(squirrel.Eq{"output_product_id": *productID})
	}
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	if offset > 0 {
		builder = builder.Offset(uint64(offset))
	}

	return repo.queryOperations(ctx, "GetOperations", builder)
}

// GetOperationByID возвращает операцию по ID.
func (repo *Repository) GetOperationByID(ctx context.Context, id int64) (*entities.Operation, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	sql, args, err := psql.Select(operationColumns...).
		From(operationsTable).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetOperationByID: build query: %w", err)
	}

	op, err := scanOperation(repo.Pool.QueryRow(ctx, sql, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.NotFound(fmt.Sprintf("Operation with id %d not found", id), err)
	}
	if err != nil {
		return nil, fmt.Errorf("GetOperationByID: %w", err)
	}
	return op, nil
}

// UpdateOperation обновляет поля таблицы operations из updates и возвращает обновленную операцию.
// Ключи updates - имена столбцов из updatableColumns; ключ "parameters" обновляет параметры операции.
func (repo *Repository) UpdateOperation(ctx context.Context, id int64, updates map[string]any) (*entities.Operation, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	builder := psql.Update(operationsTable).Where(squirrel.Eq{"id": id})
	hasColumns := false
	var parameters map[string]any
	for key, value := range updates {
		if key == "parameters" {
			params, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("UpdateOperation: parameters must be map[string]any, got %T", value)
			}
			parameters = params
			continue
		}
		convert, ok := updatableColumns[key]
		if !ok {
			return nil, fmt.Errorf("UpdateOperation: field %q cannot be updated", key)
		}
		converted, err := convert(value)
		if err != nil {
			return nil, fmt.Errorf("UpdateOperation: field %q: %w", key, err)
		}
		builder = builder.Set(key, converted)
		hasColumns = true
	}

	tx, err := repo.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("UpdateOperation: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // после Commit откат ничего не делает

	var op *entities.Operation
	if hasColumns {
		sql, args, buildErr := builder.Suffix("RETURNING " + strings.Join(operationColumns, ", ")).ToSql()
		if buildErr != nil {
			return nil, fmt.Errorf("UpdateOperation: build query: %w", buildErr)
		}
		op, err = scanOperation(tx.QueryRow(ctx, sql, args...))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("Operation with id %d not found", id), err)
		}
		if err != nil {
			return nil, wrapDBError("UpdateOperation", err)
		}
	}

	if parameters != nil {
		if err = replaceParameters(ctx, tx, id, parameters); err != nil {
			return nil, err
		}
	}

	if op == nil {
		sql, args, buildErr := psql.Select(operationColumns...).
			From(operationsTable).
			Where(squirrel.Eq{"id": id}).
			ToSql()
		if buildErr != nil {
			return nil, fmt.Errorf("UpdateOperation: build query: %w", buildErr)
		}
		op, err = scanOperation(tx.QueryRow(ctx, sql, args...))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("Operation with id %d not found", id), err)
		}
		if err != nil {
			return nil, fmt.Errorf("UpdateOperation: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("UpdateOperation: commit: %w", err)
	}
	if parameters != nil {
		op.Parameters = parameters
	}
	return op, nil
}

// DeleteOperation удаляет операцию по ID. Параметры операции удаляются каскадно.
func (repo *Repository) DeleteOperation(ctx context.Context, id int64) error {
	if repo.Pool == nil {
		return errors.New("db pool is empty")
	}

	sql, args, err := psql.Delete(operationsTable).Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("DeleteOperation: build query: %w", err)
	}

	tag, err := repo.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteOperation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.NotFound(fmt.Sprintf("Operation with id %d not found", id), nil)
	}
	return nil
}

// GetProductOperations возвращает все операции маршрута изделия в порядке номеров шагов.
func (repo *Repository) GetProductOperations(ctx context.Context, productID int64) ([]entities.Operation, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	builder := psql.Select(operationColumns...).
		From(operationsTable).
		Where(squirrel.Eq{"output_product_id": productID}).
		OrderBy("step_number", "id")

	return repo.queryOperations(ctx, "GetProductOperations", builder)
}

// insertParameters сохраняет параметры операции.
func insertParameters(_ context.Context, _ pgx.Tx, _ int64, params map[string]any) error {
	if len(params) == 0 {
		return nil
	}
	return apperrors.Validation("operation parameters are not supported yet", nil)
}

// replaceParameters заменяет параметры операции новыми.
// TODO: реализовать вместе с insertParameters, когда будет согласован формат параметров.
func replaceParameters(ctx context.Context, tx pgx.Tx, operationID int64, params map[string]any) error {
	return insertParameters(ctx, tx, operationID, params)
}

func (repo *Repository) queryOperations(ctx context.Context, op string, builder squirrel.SelectBuilder) ([]entities.Operation, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	rows, err := repo.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	operations := make([]entities.Operation, 0)
	for rows.Next() {
		item, scanErr := scanOperation(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, scanErr)
		}
		operations = append(operations, *item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	return operations, nil
}

// scanOperation читает строку со столбцами operationColumns.
func scanOperation(row pgx.Row) (*entities.Operation, error) {
	var (
		op          entities.Operation
		description *string
		stepNumber  int32
	)
	err := row.Scan(
		&op.ID,
		&op.Name,
		&description,
		&op.ClassID,
		&op.ProfessionID,
		&op.LaborCost,
		&stepNumber,
		&op.Equipment.Name,
		&op.Equipment.WorkCenter,
		&op.Materials.MaterialID,
		&op.Materials.ProductID,
	)
	if err != nil {
		return nil, err
	}
	if description != nil {
		op.Description = *description
	}
	op.StepNumber = stepNumber
	return &op, nil
}

// wrapDBError переводит ошибки ограничений PostgreSQL в ошибки приложения.
func wrapDBError(op string, err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return fmt.Errorf("%s: %w", op, err)
	}
	switch pgErr.Code {
	case "23503": // foreign_key_violation
		return apperrors.NotFound(foreignKeyMessage(pgErr.ConstraintName), err)
	case "23505": // unique_violation
		return apperrors.Conflict("Operation with this step number already exists for the product", err)
	case "23514", "23502": // check_violation, not_null_violation
		return apperrors.Validation(fmt.Sprintf("Invalid operation data: %s", pgErr.Message), err)
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}

func foreignKeyMessage(constraint string) string {
	switch constraint {
	case "operations_operation_class_id_fkey":
		return "Operation class not found"
	case "operations_profession_id_fkey":
		return "Profession not found"
	case "fk_operations_work_center":
		return "Equipment not found in the work center"
	case "fk_operations_material":
		return "Input product is not a material of the output product"
	case "operations_input_product_id_fkey":
		return "Input product not found"
	case "operations_output_product_id_fkey":
		return "Output product not found"
	default:
		return "Related entity not found"
	}
}

func nullableText(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toText(v any) (any, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	case *string:
		return t, nil
	default:
		return nil, fmt.Errorf("expected string, got %T", v)
	}
}

func toInt64(v any) (any, error) {
	switch t := v.(type) {
	case int:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	case float64:
		if t != float64(int64(t)) {
			return nil, fmt.Errorf("expected integer, got %v", t)
		}
		return int64(t), nil
	default:
		return nil, fmt.Errorf("expected integer, got %T", v)
	}
}

func toFloat64(v any) (any, error) {
	switch t := v.(type) {
	case float32:
		return float64(t), nil
	case float64:
		return t, nil
	case int:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	default:
		return nil, fmt.Errorf("expected number, got %T", v)
	}
}
