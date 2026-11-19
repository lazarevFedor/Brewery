package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// EquipmentRepository определяет доступ к хранилищу оборудования.
type EquipmentRepository interface {
	// InsertEquipment сохраняет оборудование и возвращает сохраненные данные.
	InsertEquipment(ctx context.Context, eq entities.Equipment) (*entities.Equipment, error)

	// GetEquipment возвращает список оборудования с пагинацией и необязательными фильтрами.
	GetEquipment(ctx context.Context, limit, offset int32, name, workCenter *string) ([]entities.Equipment, error)

	// GetEquipmentByKey возвращает оборудование по паре названия и цеха.
	GetEquipmentByKey(ctx context.Context, name, workCenter string) (*entities.Equipment, error)

	// UpdateEquipment обновляет переданные поля по исходному ключу и возвращает сохраненную запись.
	UpdateEquipment(ctx context.Context, name, workCenter string, updates map[string]any) (*entities.Equipment, error)

	// DeleteEquipment удаляет оборудование по ключу. При отсутствии записи возвращает ошибку.
	DeleteEquipment(ctx context.Context, name, workCenter string) error
}

// EquipmentService реализует бизнес-логику работы с оборудованием.
type EquipmentService struct {
	Repo EquipmentRepository
}

// CreateEquipment проверяет данные и сохраняет оборудование.
func (s *EquipmentService) CreateEquipment(ctx context.Context, eq entities.Equipment) (*entities.Equipment, error) {
	if err := validateEquipmentKey(eq.Name, eq.WorkCenter); err != nil {
		return nil, err
	}
	created, err := s.Repo.InsertEquipment(ctx, eq)
	if err != nil {
		return nil, equipmentRepositoryError(err)
	}
	if created == nil {
		return nil, errors.New("equipment repository returned no equipment after creation")
	}
	return created, nil
}

// GetEquipment возвращает страницу оборудования с фильтрами по названию и цеху.
func (s *EquipmentService) GetEquipment(ctx context.Context, limit, offset int32, name, workCenter *string) ([]entities.Equipment, error) {
	if limit <= 0 || offset < 0 {
		return nil, apperrors.BadRequest("invalid pagination", nil)
	}
	if name != nil && strings.TrimSpace(*name) == "" {
		return nil, apperrors.BadRequest("name filter must not be blank", nil)
	}
	if workCenter != nil && strings.TrimSpace(*workCenter) == "" {
		return nil, apperrors.BadRequest("work_center filter must not be blank", nil)
	}
	return s.Repo.GetEquipment(ctx, limit, offset, name, workCenter)
}

// GetEquipmentByKey возвращает оборудование по составному ключу.
func (s *EquipmentService) GetEquipmentByKey(ctx context.Context, name, workCenter string) (*entities.Equipment, error) {
	if err := validateEquipmentKey(name, workCenter); err != nil {
		return nil, err
	}
	eq, err := s.Repo.GetEquipmentByKey(ctx, name, workCenter)
	if err != nil {
		return nil, equipmentRepositoryError(err)
	}
	if eq == nil {
		return nil, apperrors.NotFound("equipment not found", nil)
	}
	return eq, nil
}

// UpdateEquipment проверяет и сохраняет частичное изменение оборудования.
func (s *EquipmentService) UpdateEquipment(ctx context.Context, name, workCenter string, updates map[string]any) (*entities.Equipment, error) {
	if err := validateEquipmentKey(name, workCenter); err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return nil, apperrors.BadRequest("empty equipment update", nil)
	}
	for field, value := range updates {
		if field != "name" && field != "work_center" {
			return nil, apperrors.BadRequest(fmt.Sprintf("unknown equipment field %q", field), nil)
		}
		if str, ok := value.(string); !ok || strings.TrimSpace(str) == "" {
			return nil, apperrors.BadRequest(fmt.Sprintf("%s must be a non-blank string", field), nil)
		}
	}

	eq, err := s.Repo.UpdateEquipment(ctx, name, workCenter, updates)
	if err != nil {
		return nil, equipmentRepositoryError(err)
	}
	if eq == nil {
		return nil, apperrors.NotFound("equipment not found", nil)
	}
	return eq, nil
}

// DeleteEquipment удаляет оборудование по составному ключу.
func (s *EquipmentService) DeleteEquipment(ctx context.Context, name, workCenter string) error {
	if err := validateEquipmentKey(name, workCenter); err != nil {
		return err
	}
	return equipmentRepositoryError(s.Repo.DeleteEquipment(ctx, name, workCenter))
}

func validateEquipmentKey(name, workCenter string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(workCenter) == "" {
		return apperrors.BadRequest("name and work_center are required", nil)
	}
	return nil
}

func equipmentRepositoryError(err error) error {
	if _, ok := errors.AsType[*apperrors.AppError](err); ok {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.NotFound("equipment not found", err)
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return apperrors.Conflict("equipment with this name and work_center already exists", err)
		case "23503", "23001":
			return apperrors.Conflict("equipment is used by an operation", err)
		}
	}
	return err
}
