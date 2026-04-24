package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
)

type EnumValueHandlers interface {
	CreateValue(c *gin.Context)
	GetValue(c *gin.Context)
	UpdateValue(c *gin.Context)
	DeleteValue(c *gin.Context)
}

type enumValueHandlers struct {
	uc usecase.EnumService
}

func NewEnumValueHandlers(usecase usecase.EnumService) EnumValueHandlers {
	return &enumValueHandlers{
		uc: usecase,
	}
}


func (h *enumValueHandlers) CreateValue(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	log.Debug(c.Request.Context(), "Start Create")

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read category in the request body: %v", err))
		return
	}

	log.Debug(c.Request.Context(), "Bode read Create")


	var req entities.EnumValue
	if err = easyjson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal category: %v", err))
		return
	}

	log.Debug(c.Request.Context(), "Unmarshaled Create")


	enumValueID, err := h.uc.CreateEnumValue(c.Request.Context(), req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get category: %v", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	log.Debug(c.Request.Context(), "End Create")

	log.Debug(c.Request.Context(), fmt.Sprintf("enum_value_id=%d", enumValueID))
	log.Info(
		c.Request.Context(),
		// fmt.Sprintf("action=create resource=category status=success name=%q level=%d",
		// 	req.Name, req.Level))
		fmt.Sprintf("action=create resource=category status=success value=%q", req.Value),
	)
	c.Status(http.StatusCreated)
}

func (h *enumValueHandlers) GetValue(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read category in the request body: %v", err))
		return
	}

	req := make(map[string]string)
	if err = json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal enum get data: %v", err))
		return
	}

	entityName, ok := req["entity_name"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("entity_name is empty: %v", err))
		return
	}

	fieldName, ok := req["field_name"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("field_name is empty: %v", err))
		return
	}

	valueType, ok := req["type"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("type is empty: %v", err))
		return
	}

	values, err := h.uc.GetEnumValue(c.Request.Context(), entityName, fieldName, valueType)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get enum values: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	rawBytes, err := json.Marshal(values)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal enums: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(
		c.Request.Context(),
		fmt.Sprintf(
			"action=list resource=beer status=success items=%d", len(values),
		),
	)
}

func (h *enumValueHandlers) UpdateValue(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid enum value id: %v", err))

		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read updates in the request body: %v", err))

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

	err = h.uc.UpdateEnumValue(c.Request.Context(), id, updates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update enum value: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=beer status=success id=%d", id))
	c.Status(http.StatusOK)
}

func (h *enumValueHandlers) DeleteValue(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid enum value id: %v", err))

		return
	}

	err = h.uc.DeleteEnumValue(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete enum value: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=beer status=success id=%d", id))
	c.Status(http.StatusOK)
}

