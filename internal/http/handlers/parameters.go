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

type ParametersHandlers interface {
	CreateNumericParameter(c *gin.Context)
	UpdateNumericParameter(c *gin.Context)
	DeleteNumericParameter(c *gin.Context)

	CreateEnumParameter(c *gin.Context)
	UpdateEnumParameter(c *gin.Context)
	DeleteEnumParameter(c *gin.Context)

	UpdateParameter(c *gin.Context)
	ListCategoryParameters(c *gin.Context)
	ApplyParametersToCategory(c *gin.Context)
}

type parametersHandlers struct {
	uc usecase.ParametersService
}

func NewParametersHandlers(uc usecase.ParametersService) ParametersHandlers {
	return &parametersHandlers{
		uc: uc,
	}
}

func (h *parametersHandlers) CreateNumericParameter(c *gin.Context) {
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

	var req entities.NumericParameter
	if err = easyjson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal numeric parameter: %v", err))
		return
	}

	created, err := h.uc.CreateNumeric(c.Request.Context(), &req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create numeric parameter: %v", err))
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
	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=numeric_parameter status=success id=%d", created.ID))
}

func (h *parametersHandlers) UpdateNumericParameter(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid numeric parameter id: %v", err))
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

	updated, err := h.uc.UpdateNumeric(c.Request.Context(), id, updates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update numeric parameter: %v", err))
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
	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=numeric_parameter status=success id=%d", id))
}

func (h *parametersHandlers) DeleteNumericParameter(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid numeric parameter id: %v", err))
		return
	}

	deleted, err := h.uc.DeleteNumeric(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete numeric parameter: %v", err))
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
	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=numeric_parameter status=success id=%d", id))
}

func (h *parametersHandlers) CreateEnumParameter(c *gin.Context) {
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

	var req entities.EnumParameter
	if err = easyjson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal enum parameter: %v", err))
		return
	}

	created, err := h.uc.CreateEnum(c.Request.Context(), &req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create enum parameter: %v", err))
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
	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=enum_parameter status=success id=%d", created.ID))
}

func (h *parametersHandlers) UpdateEnumParameter(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid enum parameter id: %v", err))
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

	updated, err := h.uc.UpdateEnum(c.Request.Context(), id, updates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update enum parameter: %v", err))
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
	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=enum_parameter status=success id=%d", id))
}

//nolint:funlen
func (h *parametersHandlers) UpdateParameter(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid parameter id: %v", err))
		return
	}

	typ := c.Query("type")
	if typ == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type query param is required"})
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

	switch typ {
	case "numeric":
		updated, err := h.uc.UpdateNumeric(c.Request.Context(), id, updates)
		if err != nil {
			log.Error(c.Request.Context(), fmt.Sprintf("Failed to update numeric parameter: %v", err))
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
		log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=numeric_parameter status=success id=%d", id))
	case "enum":
		updated, err := h.uc.UpdateEnum(c.Request.Context(), id, updates)
		if err != nil {
			log.Error(c.Request.Context(), fmt.Sprintf("Failed to update enum parameter: %v", err))
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
		log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=enum_parameter status=success id=%d", id))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type, must be 'numeric' or 'enum'"})
		return
	}
}

func (h *parametersHandlers) DeleteEnumParameter(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid enum parameter id: %v", err))
		return
	}

	deleted, err := h.uc.DeleteEnum(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete enum parameter: %v", err))
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
	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=enum_parameter status=success id=%d", id))
}

func (h *parametersHandlers) ListCategoryParameters(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	categoryID, err := getUintParam(c, "category_id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid enum parameter id: %v", err))
		return
	}

	paramTypeStr := c.Query("type")

	var paramType int
	switch paramTypeStr {
	case "numeric":
		paramType = entities.NumericParameterType
	case "enum":
		paramType = entities.EnumParameterType
	case "":
		paramType = entities.MissingType
	}

	log.Debug(c.Request.Context(), fmt.Sprintf("type: %s, %d", paramTypeStr, paramType))

	numeric, enum, err := h.uc.ListParameters(c.Request.Context(), categoryID, paramType)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to list parameters: %v", err))
		c.Status(http.StatusInternalServerError)
		return
	}

	result := make([]any, 0)
	for _, p := range numeric {
		result = append(result, p)
	}
	for _, p := range enum {
		result = append(result, p)
	}

	rawBytes, err := json.Marshal(result)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal response: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=list resource=parameters status=success numeric_count=%d enum_count=%d", len(numeric), len(enum)))
}

func (h *parametersHandlers) ApplyParametersToCategory(c *gin.Context) {
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

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read request body: %v", err))
		return
	}

	var req struct {
		NumericParamIDs []int `json:"numeric_param_ids"`
		EnumParamIDs    []int `json:"enum_param_ids"`
	}

	if err = json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal request: %v", err))
		return
	}

	added, err := h.uc.ApplyToCategory(c.Request.Context(), uint(categoryID), req.NumericParamIDs, req.EnumParamIDs)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to apply parameters: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Parameters applied successfully",
		"params_added": added,
	})
	log.Info(c.Request.Context(), fmt.Sprintf("action=apply resource=parameters_to_category status=success category_id=%d added=%d", categoryID, added))
}
