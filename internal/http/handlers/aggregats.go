package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
)

type AggregateHandlers interface {
	CreateAggregate(c *gin.Context)
	GetAggregates(c *gin.Context)
	UpdateAggregate(c *gin.Context)
	DeleteAggregate(c *gin.Context)
	ApplyAggregateToCategory(c *gin.Context)
}

type aggregateHandlers struct {
	uc usecase.AggregateService
}

func NewAggregateHandlers(uc usecase.AggregateService) AggregateHandlers {
	return &aggregateHandlers{
		uc: uc,
	}
}

func (h *aggregateHandlers) CreateAggregate(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read request body: %v", err))
		return
	}

	var req entities.Aggregate
	if err = easyjson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal aggregate: %v", err))
		return
	}

	created, err := h.uc.CreateAggregate(c.Request.Context(), &req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create aggregate: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawBytes, err := json.Marshal(created)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal response: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusCreated, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=aggregate status=success id=%d name=%q", created.ID, created.Name))
}

func (h *aggregateHandlers) GetAggregates(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	name := c.Query("name")

	aggregates, err := h.uc.GetAggregates(c.Request.Context(), name)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get aggregates: %v", err))
		c.Status(http.StatusInternalServerError)
		return
	}

	rawBytes, err := json.Marshal(aggregates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal response: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=list resource=aggregates status=success count=%d", len(aggregates)))
}

func (h *aggregateHandlers) UpdateAggregate(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid aggregate id: %v", err))
		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read request body: %v", err))
		return
	}

	updates := make(map[string]any)
	if err = json.Unmarshal(body, &updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty update payload"})
		return
	}

	updated, err := h.uc.UpdateAggregate(c.Request.Context(), id, updates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update aggregate: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rawBytes, err := json.Marshal(updated)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal response: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=aggregate status=success id=%d", id))
}

func (h *aggregateHandlers) DeleteAggregate(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid aggregate id: %v", err))
		return
	}

	deleted, err := h.uc.DeleteAggregate(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete aggregate: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rawBytes, err := json.Marshal(deleted)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal response: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=aggregate status=success id=%d", id))
}

func (h *aggregateHandlers) ApplyAggregateToCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	categoryIDStr := c.Param("category_id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category_id: %v", err))
		return
	}

	aggregateIDStr := c.Query("aggregate_id")
	aggregateID, err := strconv.ParseUint(aggregateIDStr, 10, 32)
	if err != nil || aggregateIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "aggregate_id query parameter is required"})
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid aggregate_id: %v", err))
		return
	}

	added, err := h.uc.ApplyAggregateToCategory(c.Request.Context(), uint(categoryID), uint(aggregateID))
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to apply aggregate: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"params_added": added,
	})
	log.Info(c.Request.Context(), fmt.Sprintf("action=apply resource=aggregate_to_category status=success category_id=%d aggregate_id=%d added=%d", categoryID, aggregateID, added))
}
