package usecase

import (
	"Brewery/internal/entities"
	"context"
)

// EquipmentRepository определяет доступ к хранилищу оборудования.
type EquipmentRepository interface {
	// InsertEquipment сохраняет оборудование и возвращает сохраненные данные.
	InsertEquipment(ctx context.Context, eq entities.Equipment) (*entities.Equipment, error)

	// GetEquipment возвращает список оборудования с пагинацией.
	GetEquipment(ctx context.Context, limit, offset int32) ([]entities.Equipment, error)

	//// GetEquipmentByID
	//GetEquipmentByID(ctx context.Context, id int64) (*entities.Equipment, error)
	//
	//// UpdateEquipment
	//UpdateEquipment(ctx context.Context, id int64, updates map[string]any) (*entities.Equipment, error)
	//
	//// DeleteEquipment
	//DeleteEquipment(ctx context.Context, id int64) error
}
