package repository

import (
	"Brewery/internal/entities"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EnumRepository interface {
	InsertEnumClass(ctx context.Context, enumClass entities.EnumClass) (uint, error)
	UpdateEnumClass(ctx context.Context, id uint, updates map[string]any) error
	DeleteEnumClassByID(ctx context.Context, id uint) error
	GetEnumClasses(ctx context.Context) ([]entities.EnumClass, error)

	InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error)
	UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error
	DeleteEnumValueByID(ctx context.Context, id uint) error
	GetEnumValues(ctx context.Context) ([]entities.EnumValue, error)
}

type EnumPostgres struct {
	Pool *pgxpool.Pool
}

func NewEnumPostgres(pool *pgxpool.Pool) *EnumPostgres {
	return &EnumPostgres{Pool: pool}
}

func (e *EnumPostgres) InsertEnumClass(ctx context.Context, enumClass entities.EnumClass) (uint, error) {
	return 0, nil
}

func (e *EnumPostgres) UpdateEnumClass(ctx context.Context, id uint, updates map[string]any) error {
	return nil
}

func (e *EnumPostgres) DeleteEnumClassByID(ctx context.Context, id uint) error {
	return nil
}

func (e *EnumPostgres) GetEnumClasses(ctx context.Context) ([]entities.EnumClass, error) {
	return nil, nil
}

func (e *EnumPostgres) InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error) {
	return 0, nil
}

func (e *EnumPostgres) UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error {
	return nil
}

func (e *EnumPostgres) DeleteEnumValueByID(ctx context.Context, id uint) error {
	return nil
}

func (e *EnumPostgres) GetEnumValues(ctx context.Context) ([]entities.EnumValue, error) {
	return nil, nil
}
