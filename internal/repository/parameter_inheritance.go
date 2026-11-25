package repository

import (
	"Brewery/internal/repository/queries"
	"context"

	"github.com/jackc/pgx/v5"
)

// toggleParameterInheritance распространяет изменение наследования в текущей транзакции.
func toggleParameterInheritance(ctx context.Context, tx pgx.Tx, isEnum bool, id int64, oldValue, newValue bool) error {
	if oldValue == newValue {
		return nil
	}

	query := queries.SelectClassesWithRangeParameter(id)
	if isEnum {
		query = queries.SelectClassesWithEnumParameter(id)
	}
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	classIDs, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return err
	}
	if len(classIDs) == 0 {
		return nil
	}

	if newValue {
		for _, classID := range classIDs {
			sql, args, err := queries.InheritParametersToChildren(classID).ToSql()
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, sql, args...); err != nil {
				return err
			}
		}
		return nil
	}

	removeQuery := queries.RemoveRangeParametersFromDescendants([]int64{id}, classIDs)
	if isEnum {
		removeQuery = queries.RemoveEnumParametersFromDescendants([]int64{id}, classIDs)
	}
	sql, args, err = removeQuery.ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, sql, args...)
	return err
}
