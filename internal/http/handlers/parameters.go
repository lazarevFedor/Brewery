package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ParametersService - сервис для управления параметрами классов.
type ParametersService interface {

	// CreateRange создает числовой параметр.
	CreateRange(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error)

	// UpdateRange обновляет числовой параметр. Обновлять можно только min_val, max_val, inheritable и unit.
	UpdateRange(ctx context.Context, id int64, updates map[string]any) (*entities.RangeParameter, error)

	// DeleteRange удаляет числовой параметр.
	DeleteRange(ctx context.Context, id int64) (*entities.RangeParameter, error)

	// ListParameters возвращает параметры. classID = 0 - без фильтра по классу.
	ListParameters(ctx context.Context, classID int64, parameterType int) ([]entities.RangeParameter, []entities.EnumParameter, error)

	// ApplyToClass применяет параметры к классу. Возвращает количество добавленных параметров.
	ApplyToClass(ctx context.Context, classID int64, rangeIDs, enumIDs []int64) (int, error)

	// ExcludeFromClass исключает параметры из класса. Возвращает количество исключенных параметров.
	ExcludeFromClass(ctx context.Context, classID int64, rangeIDs, enumIDs []int64) (int, error)
}

// ParametersHandlers - HTTP-обработчики для параметров классов.
type ParametersHandlers interface {

	// CreateRangeParameter обрабатывает POST /classes/parameters/range.
	CreateRangeParameter(c *gin.Context)

	// UpdateParameter обрабатывает PATCH /classes/parameters/{id}.
	UpdateParameter(c *gin.Context)

	// DeleteParameter обрабатывает DELETE /classes/parameters/{id}.
	DeleteParameter(c *gin.Context)

	// ListParameters обрабатывает GET /classes/parameters.
	ListParameters(c *gin.Context)

	// ApplyParametersToClass обрабатывает PATCH /classes/parameters/apply/{class_id}.
	ApplyParametersToClass(c *gin.Context)

	// ExcludeParametersFromClass обрабатывает PATCH /classes/parameters/exclude/{class_id}.
	ExcludeParametersFromClass(c *gin.Context)
}

// classParamsRequest - тело запросов apply/exclude (схемы ClassParamsApply и ClassParamsExclude).
type classParamsRequest struct {
	RangeParamIDs []int64 `json:"range_param_ids"`
	EnumParamIDs  []int64 `json:"enum_param_ids"`
}

// parameterTypes сопоставляет значение query-параметра type с типом параметра.
// numeric оставлен для совместимости со старым API.
var parameterTypes = map[string]int{
	"":        entities.MissingType,
	"range":   entities.RangeParameterType,
	"numeric": entities.RangeParameterType,
	"enum":    entities.EnumParameterType,
}

// parametersHandlers - реализация интерфейса ParametersHandlers.
type parametersHandlers struct {
	uc ParametersService
}

// NewParametersHandlers создает обработчики параметров классов.
func NewParametersHandlers(uc ParametersService) ParametersHandlers {
	return &parametersHandlers{uc: uc}
}

// CreateRangeParameter обрабатывает POST /classes/parameters/range.
func (h *parametersHandlers) CreateRangeParameter(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req entities.RangeParameter
	if !bindJSON(c, &req) {
		return
	}

	if req.FieldName == "" {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "field_name is required")
		return
	}

	if req.MinValue > req.MaxValue {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "min_val must not be greater than max_val")
		return
	}

	created, err := h.uc.CreateRange(c.Request.Context(), &req)
	if err != nil {
		handleError(c, log, "failed to create range parameter", err)
		return
	}

	log.Info(c.Request.Context(), "action=create resource=range_parameter status=success", zap.Int64("id", created.ID))
	c.JSON(http.StatusCreated, created)
}

// UpdateParameter обрабатывает PATCH /classes/parameters/{id}.
func (h *parametersHandlers) UpdateParameter(c *gin.Context) {
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

	updated, err := h.uc.UpdateRange(c.Request.Context(), id, updates)
	if err != nil {
		handleError(c, log, "failed to update range parameter", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=range_parameter status=success", zap.Int64("id", id))
	c.JSON(http.StatusOK, updated)
}

// DeleteParameter обрабатывает DELETE /classes/parameters/{id}.
func (h *parametersHandlers) DeleteParameter(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	if _, err = h.uc.DeleteRange(c.Request.Context(), id); err != nil {
		handleError(c, log, "failed to delete range parameter", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=range_parameter status=success", zap.Int64("id", id))
	c.Status(http.StatusNoContent)
}

// ListParameters обрабатывает GET /classes/parameters.
func (h *parametersHandlers) ListParameters(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classID, err := getIDQuery(c, "class_id", false)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	paramType, ok := parameterTypes[c.Query("type")]
	if !ok {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("type must be 'range' or 'enum'"))
		return
	}

	rangeParams, enumParams, err := h.uc.ListParameters(c.Request.Context(), classID, paramType)
	if err != nil {
		handleError(c, log, "failed to list parameters", err)
		return
	}

	result := make([]any, 0, len(rangeParams)+len(enumParams))
	for _, p := range rangeParams {
		result = append(result, p)
	}
	for _, p := range enumParams {
		result = append(result, p)
	}

	c.JSON(http.StatusOK, result)
}

// ApplyParametersToClass обрабатывает PATCH /classes/parameters/apply/{class_id}.
func (h *parametersHandlers) ApplyParametersToClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classID, req, ok := h.bindClassParams(c)
	if !ok {
		return
	}

	added, err := h.uc.ApplyToClass(c.Request.Context(), classID, req.RangeParamIDs, req.EnumParamIDs)
	if err != nil {
		handleError(c, log, "failed to apply parameters", err)
		return
	}

	log.Info(c.Request.Context(), "action=apply resource=parameters status=success",
		zap.Int64("class_id", classID), zap.Int("added", added))
	c.JSON(http.StatusOK, gin.H{
		"message":      "Parameters applied successfully",
		"params_added": added,
	})
}

// ExcludeParametersFromClass обрабатывает PATCH /classes/parameters/exclude/{class_id}.
func (h *parametersHandlers) ExcludeParametersFromClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classID, req, ok := h.bindClassParams(c)
	if !ok {
		return
	}

	excluded, err := h.uc.ExcludeFromClass(c.Request.Context(), classID, req.RangeParamIDs, req.EnumParamIDs)
	if err != nil {
		handleError(c, log, "failed to exclude parameters", err)
		return
	}

	log.Info(c.Request.Context(), "action=exclude resource=parameters status=success",
		zap.Int64("class_id", classID), zap.Int("excluded", excluded))
	c.JSON(http.StatusOK, gin.H{
		"message":         "Parameters excluded successfully",
		"params_excluded": excluded,
	})
}

// bindClassParams разбирает class_id из пути и список параметров из тела запросов apply/exclude.
func (h *parametersHandlers) bindClassParams(c *gin.Context) (int64, classParamsRequest, bool) {
	var req classParamsRequest

	log, ok := getLogger(c)
	if !ok {
		return 0, req, false
	}

	classID, err := getIDParam(c, "class_id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return 0, req, false
	}

	if !bindJSON(c, &req) {
		return 0, req, false
	}

	if len(req.RangeParamIDs) == 0 && len(req.EnumParamIDs) == 0 {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "range_param_ids or enum_param_ids must not be empty")
		return 0, req, false
	}

	return classID, req, true
}
