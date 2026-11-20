package operation

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
)

func loadParameters(ctx context.Context, q *queries.Queries, ids []int64) (map[int64]map[string]any, error) {
	byID := make(map[int64]map[string]any, len(ids))
	if len(ids) == 0 {
		return byID, nil
	}
	for _, id := range ids {
		byID[id] = make(map[string]any)
	}
	rows, err := q.GetOperationParameterValues(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get operation parameters: %w", err)
	}
	for _, row := range rows {
		value, decodeErr := (&entities.ParameterValueRow{ValueRaw: row.ValueRaw, ValueType: entities.EnumType(row.ValueType)}).FromRow()
		if decodeErr != nil {
			return nil, fmt.Errorf("decode parameter %q of operation %d: %w", row.FieldName, row.OperationID, decodeErr)
		}
		if _, exists := byID[row.OperationID][row.FieldName]; exists {
			return nil, fmt.Errorf("ambiguous parameter field %q for operation %d", row.FieldName, row.OperationID)
		}
		byID[row.OperationID][row.FieldName] = value.Value
	}
	return byID, nil
}

func saveParameters(ctx context.Context, q *queries.Queries, operationID int64, values []entities.ParameterValue) error {
	for _, value := range values {
		var err error
		if value.RangeParameterID != nil {
			err = q.InsertOperationRangeValue(ctx, queries.InsertOperationRangeValueParams{
				OperationID: operationID, RangeParameterID: value.RangeParameterID,
				NumericValueRaw: strconv.FormatFloat(*value.NumericValue, 'f', 3, 64),
			})
		} else {
			err = q.InsertOperationEnumValue(ctx, queries.InsertOperationEnumValueParams{
				OperationID: operationID, EnumParameterID: value.EnumParameterID, EnumValueID: value.EnumValueID,
			})
		}
		if err != nil {
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.ConstraintName == "fk_operations_parameters_enum_value" {
				return apperrors.Conflict(fmt.Sprintf("allowed value for parameter %q changed; reload and retry", value.FieldName), err)
			}
			return fmt.Errorf("save operation parameter: %w", err)
		}
	}
	return nil
}
