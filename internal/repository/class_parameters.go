package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ClassParametersPostgres управляет связями параметров с классами изделий.
type ClassParametersPostgres struct{ Pool *pgxpool.Pool }

// NewClassParametersPostgres создает репозиторий параметров классов.
func NewClassParametersPostgres(pool *pgxpool.Pool) *ClassParametersPostgres {
	return &ClassParametersPostgres{Pool: pool}
}

// GetParameters извлекает все числовые и перечисляемые параметры из базы данных и возвращает их.
//
//nolint:funlen
func (p *ClassParametersPostgres) GetParameters(
	ctx context.Context,
	classID int64,
	parameterType int,
) (
	[]entities.RangeParameter,
	[]entities.EnumClass,
	error,
) {
	if p.Pool == nil {
		return nil, nil, errors.New("pool is nil")
	}
	if classID < 0 {
		return nil, nil, errors.New("invalid class ID")
	}
	if parameterType != entities.MissingType && parameterType != entities.RangeParameterType && parameterType != entities.EnumParameterType {
		return nil, nil, errors.New("invalid parameter type")
	}
	var rangeIDs []int64
	var enumIDs []int64
	var query sq.SelectBuilder

	if classID != 0 {
		query := queries.SelectParameterIDsByClass(classID, parameterType)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, err
		}

		switch parameterType {
		case entities.MissingType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&rangeIDs, &enumIDs)
		case entities.RangeParameterType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&rangeIDs)
		case entities.EnumParameterType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&enumIDs)
		}
		if err != nil {
			return nil, nil, err
		}
	}

	var rangeParams []entities.RangeParameter
	var enumParams []entities.EnumClass

	if ((parameterType == entities.MissingType ||
		parameterType == entities.RangeParameterType) && classID == 0) || len(rangeIDs) > 0 {
		query = queries.SelectRangeParameters(rangeIDs)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, err
		}

		rows, err := p.Pool.Query(ctx, sql, args...)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var param *entities.RangeParameter
			param, err = scanRangeParameter(rows)
			if err != nil {
				return nil, nil, err
			}
			rangeParams = append(rangeParams, *param)
		}
		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
		rows.Close()
	}
	if ((parameterType == entities.MissingType ||
		parameterType == entities.EnumParameterType) && classID == 0) || len(enumIDs) > 0 {
		query = queries.SelectEnumParametersByIDs(enumIDs)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, err
		}

		rows, err := p.Pool.Query(ctx, sql, args...)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var param *entities.EnumClass
			param, err = scanEnumParameter(rows)
			if err != nil {
				return nil, nil, err
			}
			enumParams = append(enumParams, *param)
		}
		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
	}

	return rangeParams, enumParams, nil
}

// ApplyParameters добавляет параметры к классу и распространяет наследуемые параметры в одной транзакции.
func (p *ClassParametersPostgres) ApplyParameters(ctx context.Context, classID int64, rangeIDs, enumIDs []int64) (int, error) {
	if p.Pool == nil {
		return 0, errors.New("pool is nil")
	}
	if classID <= 0 {
		return 0, errors.New("invalid class ID")
	}

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, _, err := lockClassParameterIDs(ctx, tx, classID); err != nil {
		return 0, err
	}

	affected := 0
	for _, query := range []sq.SelectBuilder{
		queries.AddRangeParametersToClasses(rangeIDs, []int64{classID}),
		queries.AddEnumParametersToClasses(enumIDs, []int64{classID}),
		queries.InheritParametersToChildren(classID),
	} {
		sql, args, err := query.ToSql()
		if err != nil {
			return 0, err
		}
		var count int
		if err := tx.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
			return 0, err
		}
		affected += count
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return affected, nil
}

// ExcludeParameters исключает параметры из класса и убирает их наследуемые связи у потомков.
// Возвращает число параметров, фактически исключенных из выбранного класса.
func (p *ClassParametersPostgres) ExcludeParameters(ctx context.Context, classID int64, rangeIDs, enumIDs []int64) (int, error) {
	if p.Pool == nil {
		return 0, errors.New("pool is nil")
	}
	if classID <= 0 {
		return 0, errors.New("invalid class ID")
	}

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	linkedRangeIDs, linkedEnumIDs, err := lockClassParameterIDs(ctx, tx, classID)
	if err != nil {
		return 0, err
	}
	rangeIDs = intersectParameterIDs(rangeIDs, linkedRangeIDs)
	enumIDs = intersectParameterIDs(enumIDs, linkedEnumIDs)

	inheritedRangeIDs, err := getInheritableParameterIDs(ctx, tx, rangeIDs, false)
	if err != nil {
		return 0, err
	}
	inheritedEnumIDs, err := getInheritableParameterIDs(ctx, tx, enumIDs, true)
	if err != nil {
		return 0, err
	}

	for _, query := range []sq.SelectBuilder{
		queries.RemoveRangeParametersFromClasses(rangeIDs, []int64{classID}),
		queries.RemoveEnumParametersFromClasses(enumIDs, []int64{classID}),
		queries.RemoveRangeParametersFromDescendants(inheritedRangeIDs, []int64{classID}),
		queries.RemoveEnumParametersFromDescendants(inheritedEnumIDs, []int64{classID}),
	} {
		sql, args, err := query.ToSql()
		if err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(rangeIDs) + len(enumIDs), nil
}

func lockClassParameterIDs(ctx context.Context, tx pgx.Tx, classID int64) ([]int64, []int64, error) {
	sql, args, err := queries.SelectParameterIDsByClass(classID, entities.MissingType).Suffix("FOR UPDATE").ToSql()
	if err != nil {
		return nil, nil, err
	}
	var rangeIDs, enumIDs []int64
	if err := tx.QueryRow(ctx, sql, args...).Scan(&rangeIDs, &enumIDs); err != nil {
		return nil, nil, fmt.Errorf("lock parameters of class %d: %w", classID, err)
	}
	return rangeIDs, enumIDs, nil
}

func intersectParameterIDs(requested, linked []int64) []int64 {
	requestedSet := make(map[int64]struct{}, len(requested))
	for _, id := range requested {
		requestedSet[id] = struct{}{}
	}
	var result []int64
	for _, id := range linked {
		if _, ok := requestedSet[id]; ok {
			result = append(result, id)
			delete(requestedSet, id)
		}
	}
	return result
}

func getInheritableParameterIDs(ctx context.Context, tx pgx.Tx, ids []int64, isEnum bool) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := queries.SelectInheritableRangeParameterIDs(ids)
	if isEnum {
		query = queries.SelectInheritableEnumParameterIDs(ids)
	}
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[int64])
}
