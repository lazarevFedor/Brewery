package repository

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// ValidateParameterAssignments verifies assigned values against bounds and class links in the write transaction.
func ValidateParameterAssignments(ctx context.Context, q *queries.Queries, classID int64, values []entities.ParameterValue) error {
	if len(values) == 0 {
		return nil
	}
	// Lock bounds before the class, matching the order used when changing a range and its inheritance.
	rangeValues := make([]entities.ParameterValue, 0, len(values))
	for _, value := range values {
		if value.RangeParameterID == nil {
			continue
		}
		if value.NumericValue == nil || value.EnumParameterID != nil || value.EnumValueID != nil {
			return errors.New("invalid numeric parameter assignment")
		}
		rangeValues = append(rangeValues, value)
	}
	slices.SortFunc(rangeValues, func(a, b entities.ParameterValue) int {
		return cmp.Compare(*a.RangeParameterID, *b.RangeParameterID)
	})
	for _, value := range rangeValues {
		_, err := q.LockRangeParameterWithValue(ctx, queries.LockRangeParameterWithValueParams{
			RangeParameterID: *value.RangeParameterID,
			NumericValueRaw:  strconv.FormatFloat(*value.NumericValue, 'f', 3, 64),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.Conflict(fmt.Sprintf("range for parameter %q changed; reload and retry", value.FieldName), err)
		}
		if err != nil {
			return fmt.Errorf("lock range for parameter %q: %w", value.FieldName, err)
		}
	}

	// Hold the class stable until its parameter values have been saved.
	class, err := q.LockClassParameterIDs(ctx, classID)
	if err != nil {
		return fmt.Errorf("lock class parameters: %w", err)
	}
	for _, value := range values {
		if value.RangeParameterID != nil {
			if !slices.Contains(class.NumericParameterIds, *value.RangeParameterID) {
				return apperrors.Conflict("class parameters changed; reload the product and retry", nil)
			}
		} else {
			if value.EnumParameterID == nil || value.EnumValueID == nil || value.NumericValue != nil {
				return errors.New("invalid enum parameter assignment")
			}
			if !slices.Contains(class.EnumParameterIds, *value.EnumParameterID) {
				return apperrors.Conflict("class parameters changed; reload the product and retry", nil)
			}
		}
	}
	return nil
}
