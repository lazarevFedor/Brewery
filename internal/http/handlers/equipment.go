package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// EquipmentService определяет интерфейс сервисного слоя для работы с оборудованием.
type EquipmentService interface {
	// CreateEquipment создает новую единицу оборудования.
	CreateEquipment(ctx context.Context, eq entities.Equipment) (*entities.Equipment, error)

	// GetEquipment возвращает список оборудования с учетом пагинации.
	GetEquipment(ctx context.Context, limit, offset int32) ([]entities.Equipment, error)

	// GetEquipmentByName возвращает оборудование по названию.
	GetEquipmentByName(ctx context.Context, name string) (*entities.Equipment, error)

	// UpdateEquipment обновляет поля оборудования по названию.
	UpdateEquipment(ctx context.Context, name string, updates map[string]any) (*entities.Equipment, error)

	// DeleteEquipment удаляет оборудование по названию.
	DeleteEquipment(ctx context.Context, name string) error
}

// EquipmentHandlers определяет HTTP-обработчики для оборудования.
type EquipmentHandlers interface {
	// ListEquipment обрабатывает GET /equipment.
	ListEquipment(c *gin.Context)

	// CreateEquipment обрабатывает POST /equipment.
	CreateEquipment(c *gin.Context)

	// GetEquipment обрабатывает GET /equipment/{name}.
	GetEquipment(c *gin.Context)

	// UpdateEquipment обрабатывает PATCH /equipment/{name}.
	UpdateEquipment(c *gin.Context)

	// DeleteEquipment обрабатывает DELETE /equipment/{name}.
	DeleteEquipment(c *gin.Context)
}

// equipmentResponse - представление оборудования в API.
type equipmentResponse struct {
	Name       string `json:"name"`
	WorkCenter string `json:"work_center"`
}

// equipmentCreateRequest - тело запроса на создание оборудования.
type equipmentCreateRequest struct {
	Name       string `json:"name"`
	WorkCenter string `json:"work_center"`
}

func (r equipmentCreateRequest) toEntity() entities.Equipment {
	return entities.Equipment{
		Name:       r.Name,
		WorkCenter: r.WorkCenter,
	}
}

func newEquipmentResponse(eq entities.Equipment) equipmentResponse {
	return equipmentResponse{
		Name:       eq.Name,
		WorkCenter: eq.WorkCenter,
	}
}

func newEquipmentListResponse(list []entities.Equipment) []equipmentResponse {
	resp := make([]equipmentResponse, 0, len(list))
	for _, eq := range list {
		resp = append(resp, newEquipmentResponse(eq))
	}
	return resp
}

type equipmentHandlers struct {
	uc EquipmentService
}

// NewEquipmentHandlers создает новый экземпляр EquipmentHandlers.
func NewEquipmentHandlers(uc EquipmentService) EquipmentHandlers {
	return &equipmentHandlers{uc: uc}
}

// ListEquipment обрабатывает GET /equipment.
func (h *equipmentHandlers) ListEquipment(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	items, err := h.uc.GetEquipment(c.Request.Context(), limit, offset)
	if err != nil {
		handleError(c, log, "failed to list equipment", err)
		return
	}

	c.JSON(http.StatusOK, newEquipmentListResponse(items))
}

// CreateEquipment обрабатывает POST /equipment.
func (h *equipmentHandlers) CreateEquipment(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req equipmentCreateRequest
	if !bindJSON(c, &req) {
		return
	}

	if req.Name == "" || req.WorkCenter == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("name and work_center are required"))
		return
	}

	created, err := h.uc.CreateEquipment(c.Request.Context(), req.toEntity())
	if err != nil {
		handleError(c, log, "failed to create equipment", err)
		return
	}

	c.JSON(http.StatusCreated, newEquipmentResponse(*created))
}

// GetEquipment обрабатывает GET /equipment/{name}.
func (h *equipmentHandlers) GetEquipment(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	name := c.Param("name")
	if name == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("name is required"))
		return
	}

	eq, err := h.uc.GetEquipmentByName(c.Request.Context(), name)
	if err != nil {
		handleError(c, log, "failed to get equipment", err)
		return
	}

	c.JSON(http.StatusOK, newEquipmentResponse(*eq))
}

// UpdateEquipment обрабатывает PATCH /equipment/{name}.
func (h *equipmentHandlers) UpdateEquipment(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	name := c.Param("name")
	if name == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("name is required"))
		return
	}

	updates, ok := bindUpdates(c)
	if !ok {
		return
	}

	updated, err := h.uc.UpdateEquipment(c.Request.Context(), name, updates)
	if err != nil {
		handleError(c, log, "failed to update equipment", err)
		return
	}

	c.JSON(http.StatusOK, newEquipmentResponse(*updated))
}

// DeleteEquipment обрабатывает DELETE /equipment/{name}.
func (h *equipmentHandlers) DeleteEquipment(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	name := c.Param("name")
	if name == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("name is required"))
		return
	}

	if err := h.uc.DeleteEquipment(c.Request.Context(), name); err != nil {
		handleError(c, log, "failed to delete equipment", err)
		return
	}

	c.Status(http.StatusNoContent)
}
