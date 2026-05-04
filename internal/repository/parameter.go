// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ParameterRepository определяет интерфейс для работы с параметрами, включая числовые и перечисляемые параметры.
type ParameterRepository interface {
	// InsertNumericParameter добавляет новый числовой параметр в базу данных и возвращает его.
	InsertNumericParameter(param *entities.NumericParameter) (*entities.NumericParameter, error)

	// UpdateNumericParameter обновляет существующий числовой параметр в базе данных и возвращает его.
	UpdateNumericParameter(id uint, updates map[string]any) (*entities.NumericParameter, error)

	// DeleteNumericParameter удаляет числовой параметр из базы данных и возвращает его.
	DeleteNumericParameter(id uint) (*entities.NumericParameter, error)

	// InsertEnumParameter добавляет новый перечисляемый параметр в базу данных и возвращает его.
	InsertEnumParameter(param *entities.EnumParameter) (*entities.EnumParameter, error)

	// UpdateEnumParameter обновляет существующий перечисляемый параметр в базе данных и возвращает его.
	UpdateEnumParameter(id uint, updates map[string]any) (*entities.EnumParameter, error)

	// DeleteEnumParameter удаляет перечисляемый параметр из базы данных и возвращает его.
	DeleteEnumParameter(id uint) (*entities.EnumParameter, error)

	// GetParameters извлекает все числовые и перечисляемые параметры из базы данных и возвращает их.
	GetParameters(categoryID uint) ([]entities.NumericParameter, []entities.EnumParameter, error)

	// ApplyParameters применяет числовые и перечисляемые параметры к категории и возвращает результат применения.
	ApplyParameters(categoryID uint, numericParameters, enumParameters []int) (int, error)
}

// ParameterPostgres реализует интерфейс ParameterRepository для работы с параметрами в базе данных PostgreSQL, используя пул соединений pgxpool.
type ParameterPostgres struct {
	Pool *pgxpool.Pool
}

// NewParameterRepository создает новый экземпляр ParameterRepository с предоставленным пулом соединений к базе данных.
func NewParameterRepository(pool *pgxpool.Pool) ParameterRepository {
	return &ParameterPostgres{Pool: pool}
}

// NewParameterPostgres создает новый экземпляр ParameterPostgres с предоставленным пулом соединений к базе данных.
func NewParameterPostgres(pool *pgxpool.Pool) *ParameterPostgres {
	return &ParameterPostgres{Pool: pool}
}

// InsertNumericParameter добавляет новый числовой параметр в базу данных и возвращает его.
func (p *ParameterPostgres) InsertNumericParameter(param *entities.NumericParameter) (*entities.NumericParameter, error) {
	return nil, nil
}

// UpdateNumericParameter обновляет существующий числовой параметр в базе данных и возвращает его.
func (p *ParameterPostgres) UpdateNumericParameter(id uint, updates map[string]any) (*entities.NumericParameter, error) {
	return nil, nil
}

// DeleteNumericParameter удаляет числовой параметр из базы данных и возвращает его.
func (p *ParameterPostgres) DeleteNumericParameter(id uint) (*entities.NumericParameter, error) {
	return nil, nil
}

// InsertEnumParameter добавляет новый перечисляемый параметр в базу данных и возвращает его.
func (p *ParameterPostgres) InsertEnumParameter(param *entities.EnumParameter) (*entities.EnumParameter, error) {
	return nil, nil
}

// UpdateEnumParameter обновляет существующий перечисляемый параметр в базе данных и возвращает его.
func (p *ParameterPostgres) UpdateEnumParameter(id uint, updates map[string]any) (*entities.EnumParameter, error) {
	return nil, nil
}

// DeleteEnumParameter удаляет перечисляемый параметр из базы данных и возвращает его.
func (p *ParameterPostgres) DeleteEnumParameter(id uint) (*entities.EnumParameter, error) {
	return nil, nil
}

// GetParameters извлекает все числовые и перечисляемые параметры из базы данных и возвращает их.
func (p *ParameterPostgres) GetParameters(categoryID uint) ([]entities.NumericParameter, []entities.EnumParameter, error) {
	return nil, nil, nil
}

// ApplyParameters применяет числовые и перечисляемые параметры к категории и возвращает результат применения.
func (p *ParameterPostgres) ApplyParameters(categoryID uint, numericParameters, enumParameters []int) (int, error) {
	return 0, nil
}
