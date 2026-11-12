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

	// GetEquipmentByID возвращает оборудование по идентификатору.
	GetEquipmentByID(ctx context.Context, id int64) (*entities.Equipment, error)

	// UpdateEquipment обновляет поля оборудования.
	UpdateEquipment(ctx context.Context, id int64, updates map[string]any) (*entities.Equipment, error)

	// DeleteEquipment удаляет оборудование по идентификатору.
	DeleteEquipment(ctx context.Context, id int64) error
}

// EquipmentHandlers определяет HTTP-обработчики для оборудования.
type EquipmentHandlers interface {
	// ListEquipment обрабатывает GET /equipment.
	ListEquipment(c *gin.Context)

	// CreateEquipment обрабатывает POST /equipment.
	CreateEquipment(c *gin.Context)

	// GetEquipment обрабатывает GET /equipment/{id}.
	GetEquipment(c *gin.Context)

	// UpdateEquipment обрабатывает PATCH /equipment/{id}.
	UpdateEquipment(c *gin.Context)

	// DeleteEquipment обрабатывает DELETE /equipment/{id}.
	DeleteEquipment(c *gin.Context)
}

// equipmentResponse - представление оборудования в API.
type equipmentResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// equipmentCreateRequest - тело запроса на создание оборудования.
type equipmentCreateRequest struct {
	Name string `json:"name"`
}

func (r equipmentCreateRequest) toEntity() entities.Equipment {
	return entities.Equipment{
		Name: r.Name,
	}
}

func newEquipmentResponse(eq entities.Equipment) equipmentResponse {
	return equipmentResponse{
		ID:   eq.ID,
		Name: eq.Name,
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

	if req.Name == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("name is required"))
		return
	}

	created, err := h.uc.CreateEquipment(c.Request.Context(), req.toEntity())
	if err != nil {
		handleError(c, log, "failed to create equipment", err)
		return
	}

	c.JSON(http.StatusCreated, newEquipmentResponse(*created))
}

// GetEquipment обрабатывает GET /equipment/{id}.
func (h *equipmentHandlers) GetEquipment(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	eq, err := h.uc.GetEquipmentByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, log, "failed to get equipment", err)
		return
	}

	c.JSON(http.StatusOK, newEquipmentResponse(*eq))
}

// UpdateEquipment обрабатывает PATCH /equipment/{id}.
func (h *equipmentHandlers) UpdateEquipment(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	updates, ok := bindUpdates(c)
	if !ok {
		return
	}

	updated, err := h.uc.UpdateEquipment(c.Request.Context(), id, updates)
	if err != nil {
		handleError(c, log, "failed to update equipment", err)
		return
	}

	c.JSON(http.StatusOK, newEquipmentResponse(*updated))
}

// DeleteEquipment обрабатывает DELETE /equipment/{id}.
func (h *equipmentHandlers) DeleteEquipment(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	if err := h.uc.DeleteEquipment(c.Request.Context(), id); err != nil {
		handleError(c, log, "failed to delete equipment", err)
		return
	}

	c.Status(http.StatusNoContent)
}
