package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
)

type EnumHandlers interface {
	CreateEnum(c *gin.Context)
	GetEnum(c *gin.Context)
	UpdateEnum(c *gin.Context)
	DeleteEnum(c *gin.Context)

	CreateValue(c *gin.Context)
	GetValue(c *gin.Context)
	UpdateValue(c *gin.Context)
	DeleteValue(c *gin.Context)
}

type enumHandlers struct {
	uc usecase.EnumService
}

func NewEnumHandlers(usecase usecase.EnumService) EnumHandlers {
	return &enumHandlers{
		uc: usecase,
	}
}

func (h *enumHandlers) CreateEnum(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read enum class in the request body: %v", err))
		return
	}

	var req entities.EnumClass
	if err = easyjson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal enum class: %v", err))
		return
	}

	enumID, err := h.uc.CreateEnum(c.Request.Context(), req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create enum class: %v", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "enum class not found"})
		return
	}

	log.Debug(c.Request.Context(), fmt.Sprintf("enumID=%d", enumID))
	log.Info(
		c.Request.Context(),
		// fmt.Sprintf("action=create resource=category status=success name=%q level=%d",
		// 	req.Name, req.Level))
		fmt.Sprintf("action=create resource=category status=success name=%q", req.EntityName),
	)
	c.Status(http.StatusCreated)
}

func (h *enumHandlers) GetEnum(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	entityName := c.Query("entity_name")
	if entityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), errors.New("entity_name is empty").Error())
		return
	}

	fieldName := c.Query("field_name")
	if fieldName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), errors.New("field_name is empty").Error())
		return
	}

	enums, err := h.uc.GetEnum(c.Request.Context(), entityName, fieldName)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get enum classes: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	rawBytes, err := json.Marshal(enums)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal enum class: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(
		c.Request.Context(),
		fmt.Sprintf(
			"action=list resource=beer status=success items=%d", len(enums),
		),
	)
}

func (h *enumHandlers) UpdateEnum(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid enum id: %v", err))

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

	err = h.uc.UpdateEnum(c.Request.Context(), id, updates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update beer: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=beer status=success id=%d", id))
	c.Status(http.StatusOK)
}

func (h *enumHandlers) DeleteEnum(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid enum id: %v", err))

		return
	}

	err = h.uc.DeleteEnum(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete enum: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=beer status=success id=%d", id))
	c.Status(http.StatusOK)
}



func (h *enumHandlers) CreateValue(c *gin.Context) {
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
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal value: %v", err))
		return
	}

	if req.EnumClassID == 0{
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), errors.New("class_id is empty").Error())
		return
	}
	
	log.Debug(c.Request.Context(), "Unmarshaled Create")


	enumValueID, err := h.uc.CreateEnumValue(c.Request.Context(), req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create enum value: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "value not created"})
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

func (h *enumHandlers) GetValue(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	entityName := c.Query("entity_name")
	fieldName := c.Query("field_name")
	valueType := c.Query("enum_type")

	values, err := h.uc.GetEnumValue(c.Request.Context(), entityName, fieldName, entities.EnumType(valueType))
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

func (h *enumHandlers) UpdateValue(c *gin.Context) {
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

func (h *enumHandlers) DeleteValue(c *gin.Context) {
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

