// Package equipment хранит оборудование в таблице work_centers.
package equipment

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"

	"github.com/jackc/pgx/v5"
)

// Repository выполняет запросы к справочнику оборудования.
type Repository struct {
	Query *queries.Queries
}

func fromWorkCenter(row queries.WorkCenter) entities.Equipment {
	return entities.Equipment{Name: row.EquipmentName, WorkCenter: row.WorkCenterName}
}

// InsertEquipment добавляет оборудование в указанный цех.
func (r *Repository) InsertEquipment(ctx context.Context, eq entities.Equipment) (*entities.Equipment, error) {
	row, err := r.Query.InsertEquipment(ctx, queries.InsertEquipmentParams{
		EquipmentName: eq.Name, WorkCenterName: eq.WorkCenter,
	})
	if err != nil {
		return nil, err
	}
	return new(fromWorkCenter(row)), nil
}

// GetEquipment возвращает страницу оборудования с фильтрами по названию и цеху.
func (r *Repository) GetEquipment(ctx context.Context, limit, offset int32, name, workCenter *string) ([]entities.Equipment, error) {
	rows, err := r.Query.ListEquipment(ctx, queries.ListEquipmentParams{
		NameFilter: name, CenterFilter: workCenter,
		PageLimit: limit, PageOffset: offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]entities.Equipment, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromWorkCenter(row))
	}
	return items, nil
}

// GetEquipmentByKey находит оборудование по названию и цеху.
func (r *Repository) GetEquipmentByKey(ctx context.Context, name, workCenter string) (*entities.Equipment, error) {
	row, err := r.Query.GetEquipmentByKey(ctx, queries.GetEquipmentByKeyParams{
		EquipmentName: name, WorkCenterName: workCenter,
	})
	if err != nil {
		return nil, err
	}
	return new(fromWorkCenter(row)), nil
}

// UpdateEquipment изменяет переданные поля записи по исходному ключу.
func (r *Repository) UpdateEquipment(ctx context.Context, name, workCenter string, updates map[string]any) (*entities.Equipment, error) {
	params := queries.UpdateEquipmentParams{EquipmentName: name, WorkCenterName: workCenter}
	if value, ok := updates["name"].(string); ok {
		params.NewName = &value
	}
	if value, ok := updates["work_center"].(string); ok {
		params.NewCenter = &value
	}
	row, err := r.Query.UpdateEquipment(ctx, params)
	if err != nil {
		return nil, err
	}
	return new(fromWorkCenter(row)), nil
}

// DeleteEquipment удаляет оборудование по названию и цеху.
func (r *Repository) DeleteEquipment(ctx context.Context, name, workCenter string) error {
	count, err := r.Query.DeleteEquipment(ctx, queries.DeleteEquipmentParams{
		EquipmentName: name, WorkCenterName: workCenter,
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
