package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// classParamsRequest - тело запросов apply/exclude (схемы ClassParamsApply и ClassParamsExclude).
type classParamsRequest struct {
	RangeParamIDs []int64 `json:"range_param_ids"`
	EnumParamIDs  []int64 `json:"enum_param_ids"`
}

// parameterTypes сопоставляет значение query-параметра type с типом параметра.
var parameterTypes = map[string]int{
	"":      entities.MissingType,
	"range": entities.RangeParameterType,
	"enum":  entities.EnumParameterType,
}

// ListParameters обрабатывает GET /classes/parameters.
func (h *classHandlers) ListParameters(c *gin.Context) {
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
		result = append(result, enumParameterFromEntity(p))
	}

	c.JSON(http.StatusOK, result)
}

// ApplyParametersToClass обрабатывает PATCH /classes/parameters/apply/{id}.
func (h *classHandlers) ApplyParametersToClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classID, req, ok := h.bindClassParams(c)
	if !ok {
		return
	}

	added, err := h.uc.ApplyParametersToClass(c.Request.Context(), classID, req.RangeParamIDs, req.EnumParamIDs)
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

// ExcludeParametersFromClass обрабатывает PATCH /classes/parameters/exclude/{id}.
func (h *classHandlers) ExcludeParametersFromClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classID, req, ok := h.bindClassParams(c)
	if !ok {
		return
	}

	excluded, err := h.uc.ExcludeParametersFromClass(c.Request.Context(), classID, req.RangeParamIDs, req.EnumParamIDs)
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

// bindClassParams разбирает id из пути и список параметров из тела запросов apply/exclude.
func (h *classHandlers) bindClassParams(c *gin.Context) (int64, classParamsRequest, bool) {
	var req classParamsRequest

	log, ok := getLogger(c)
	if !ok {
		return 0, req, false
	}

	classID, err := getIDParam(c, "id")
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
