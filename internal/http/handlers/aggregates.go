package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AggregateService определяет интерфейс для работы с агрегатами.
type AggregateService interface {

	// CreateAggregate создает новый агрегат.
	CreateAggregate(ctx context.Context, aggregate *entities.Aggregate) (*entities.Aggregate, error)

	// GetAggregates получает список агрегатов, фильтруя по имени.
	GetAggregates(ctx context.Context, name string) ([]entities.Aggregate, error)

	// UpdateAggregate обновляет существующий агрегат по его ID.
	UpdateAggregate(ctx context.Context, id uint, updates map[string]any) (*entities.Aggregate, error)

	// DeleteAggregate удаляет агрегат по его ID.
	DeleteAggregate(ctx context.Context, id uint) (*entities.Aggregate, error)

	// ApplyAggregateToClass применяет агрегат к классу, возвращая количество добавленных параметров.
	ApplyAggregateToClass(ctx context.Context, classID, aggregateID uint) (int, error)
}

// AggregateHandlers - HTTP-обработчики для агрегатов параметров.
type AggregateHandlers interface {

	// CreateAggregate обрабатывает POST /aggregates.
	CreateAggregate(c *gin.Context)

	// GetAggregates обрабатывает GET /aggregates.
	GetAggregates(c *gin.Context)

	// UpdateAggregate обрабатывает PATCH /aggregates/{id}.
	UpdateAggregate(c *gin.Context)

	// DeleteAggregate обрабатывает DELETE /aggregates/{id}.
	DeleteAggregate(c *gin.Context)

	// ApplyAggregateToClass обрабатывает PATCH /aggregates/apply/{class_id}.
	ApplyAggregateToClass(c *gin.Context)
}

// aggregateHandlers - реализация интерфейса AggregateHandlers.
type aggregateHandlers struct {
	uc AggregateService
}

// NewAggregateHandlers создает обработчики агрегатов.
func NewAggregateHandlers(uc AggregateService) AggregateHandlers {
	return &aggregateHandlers{uc: uc}
}

// CreateAggregate обрабатывает POST /aggregates.
func (h *aggregateHandlers) CreateAggregate(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req entities.Aggregate
	if !bindJSON(c, &req) {
		return
	}

	if req.Name == "" {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "Aggregate name is required")
		return
	}

	created, err := h.uc.CreateAggregate(c.Request.Context(), &req)
	if err != nil {
		handleError(c, log, "failed to create aggregate", err)
		return
	}

	log.Info(c.Request.Context(), "action=create resource=aggregate status=success", zap.Uint("id", created.ID))
	c.JSON(http.StatusCreated, created)
}

// GetAggregates обрабатывает GET /aggregates.
func (h *aggregateHandlers) GetAggregates(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	aggregates, err := h.uc.GetAggregates(c.Request.Context(), c.Query("name"))
	if err != nil {
		handleError(c, log, "failed to get aggregates", err)
		return
	}

	c.JSON(http.StatusOK, nonNil(aggregates))
}

// UpdateAggregate обрабатывает PATCH /aggregates/{id}.
func (h *aggregateHandlers) UpdateAggregate(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	updates, ok := bindUpdates(c)
	if !ok {
		return
	}

	updated, err := h.uc.UpdateAggregate(c.Request.Context(), toUint(id), updates)
	if err != nil {
		handleError(c, log, "failed to update aggregate", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=aggregate status=success", zap.Int64("id", id))
	c.JSON(http.StatusOK, updated)
}

// DeleteAggregate обрабатывает DELETE /aggregates/{id}.
func (h *aggregateHandlers) DeleteAggregate(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	if _, err = h.uc.DeleteAggregate(c.Request.Context(), toUint(id)); err != nil {
		handleError(c, log, "failed to delete aggregate", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=aggregate status=success", zap.Int64("id", id))
	c.Status(http.StatusNoContent)
}

// ApplyAggregateToClass обрабатывает PATCH /aggregates/apply/{class_id}?aggregate_id=.
func (h *aggregateHandlers) ApplyAggregateToClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classID, err := getIDParam(c, "class_id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	aggregateID, err := getIDQuery(c, "aggregate_id", true)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	added, err := h.uc.ApplyAggregateToClass(c.Request.Context(), toUint(classID), toUint(aggregateID))
	if err != nil {
		handleError(c, log, "failed to apply aggregate", err)
		return
	}

	log.Info(c.Request.Context(), "action=apply resource=aggregate status=success",
		zap.Int64("class_id", classID), zap.Int64("aggregate_id", aggregateID), zap.Int("added", added))
	c.JSON(http.StatusOK, gin.H{"params_added": added})
}
