package repository

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RangeParametersPostgres хранит числовые параметры в PostgreSQL.
type RangeParametersPostgres struct {
	Pool *pgxpool.Pool
}

// NewRangeParametersPostgres создает репозиторий числовых параметров.
func NewRangeParametersPostgres(pool *pgxpool.Pool) *RangeParametersPostgres {
	return &RangeParametersPostgres{Pool: pool}
}

// InsertRangeParameter добавляет числовой параметр и возвращает его с присвоенным ID.
func (p *RangeParametersPostgres) InsertRangeParameter(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error) {
	if p.Pool == nil {
		return nil, errors.New("pool is nil")
	}
	if param == nil {
		return nil, errors.New("parameter is nil")
	}

	sql, args, err := queries.InsertRangeParameter(param).ToSql()
	if err != nil {
		return nil, err
	}
	if err := p.Pool.QueryRow(ctx, sql, args...).Scan(&param.ID); err != nil {
		return nil, rangeParameterWriteError(err)
	}
	return param, nil
}

// GetRangeParameters возвращает числовые параметры с необязательным точным фильтром field_name.
func (p *RangeParametersPostgres) GetRangeParameters(ctx context.Context, fieldName string) ([]entities.RangeParameter, error) {
	if p.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	sql, args, err := queries.SelectRangeParametersByField(fieldName).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parameters := make([]entities.RangeParameter, 0)
	for rows.Next() {
		parameter, err := scanRangeParameter(rows)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, *parameter)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return parameters, nil
}

// UpdateRangeParameter обновляет параметр и его наследование в одной транзакции.
func (p *RangeParametersPostgres) UpdateRangeParameter(ctx context.Context, id int64, updates map[string]any) (*entities.RangeParameter, error) {
	if p.Pool == nil {
		return nil, errors.New("pool is nil")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	// Проверка должна учитывать значения, сохранённые во время ожидания блокировки параметра.
	tx, err := p.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if lockErr := LockParameterDefinitions(ctx, tx); lockErr != nil {
		return nil, lockErr
	}

	sql, args, err := queries.SelectRangeParameters([]int64{id}).Suffix("FOR UPDATE").ToSql()
	if err != nil {
		return nil, err
	}
	oldParam, err := scanRangeParameter(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, rangeParameterWriteError(err)
	}

	sql, args, err = queries.UpdateRangeParameter(id, updates).ToSql()
	if err != nil {
		return nil, err
	}
	param, err := scanRangeParameter(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, rangeParameterWriteError(err)
	}
	if param.MinValue > oldParam.MinValue || param.MaxValue < oldParam.MaxValue {
		sql, args, err := queries.RangeParameterHasValuesOutsideBounds(id).ToSql()
		if err != nil {
			return nil, err
		}
		var outsideBounds bool
		if err := tx.QueryRow(ctx, sql, args...).Scan(&outsideBounds); err != nil {
			return nil, fmt.Errorf("check saved range parameter values: %w", err)
		}
		if outsideBounds {
			return nil, apperrors.Validation("new range excludes saved product or operation values", nil)
		}
	}
	if err := toggleParameterInheritance(ctx, tx, false, id, oldParam.Inheritable, param.Inheritable); err != nil {
		return nil, ParameterFieldNameError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return param, nil
}

// DeleteRangeParameter удаляет числовой параметр и возвращает удаленную запись.
func (p *RangeParametersPostgres) DeleteRangeParameter(ctx context.Context, id int64) (*entities.RangeParameter, error) {
	if p.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	sql, args, err := queries.DeleteRangeParameter(id).ToSql()
	if err != nil {
		return nil, err
	}
	param, err := scanRangeParameter(p.Pool.QueryRow(ctx, sql, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.NotFound("Range parameter not found", err)
	}
	if err != nil {
		return nil, parameterDeleteError(err, "Range parameter")
	}
	return param, nil
}

func rangeParameterWriteError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.NotFound("Range parameter not found", err)
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch {
		case pgErr.ConstraintName == "chk_min_lower_max":
			return apperrors.Validation("min_val must be less than or equal to max_val", err)
		case pgErr.Code == "22003":
			return apperrors.BadRequest("range parameter bound is out of range", err)
		}
	}
	return err
}

func scanRangeParameter(row pgx.Row) (*entities.RangeParameter, error) {
	var param entities.RangeParameter
	if err := row.Scan(&param.ID, &param.MinValue, &param.MaxValue, &param.FieldName, &param.Inheritable, &param.Unit); err != nil {
		return nil, fmt.Errorf("scan range parameter: %w", err)
	}
	return &param, nil
}
