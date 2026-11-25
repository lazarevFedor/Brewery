package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RangeParametersService определяет операции с числовыми параметрами.
type RangeParametersService interface {
	// CreateRangeParameter создает числовой параметр.
	CreateRangeParameter(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error)

	// ListRangeParameters возвращает числовые параметры с необязательным фильтром по имени поля.
	ListRangeParameters(ctx context.Context, fieldName string) ([]entities.RangeParameter, error)

	// UpdateRangeParameter обновляет числовой параметр. Обновлять можно только min_val, max_val, inheritable и unit.
	UpdateRangeParameter(ctx context.Context, id int64, updates map[string]any) (*entities.RangeParameter, error)

	// DeleteRangeParameter удаляет числовой параметр.
	DeleteRangeParameter(ctx context.Context, id int64) (*entities.RangeParameter, error)
}

// RangeParametersHandlers определяет HTTP-обработчики числовых параметров.
type RangeParametersHandlers interface {
	// CreateRangeParameter обрабатывает POST /range.
	CreateRangeParameter(c *gin.Context)

	// ListRangeParameters обрабатывает GET /range.
	ListRangeParameters(c *gin.Context)

	// UpdateRangeParameter обрабатывает PATCH /range/{id}.
	UpdateRangeParameter(c *gin.Context)

	// DeleteRangeParameter обрабатывает DELETE /range/{id}.
	DeleteRangeParameter(c *gin.Context)
}

type rangeParametersHandlers struct {
	uc RangeParametersService
}

// NewRangeParametersHandlers создает обработчики числовых параметров.
func NewRangeParametersHandlers(uc RangeParametersService) RangeParametersHandlers {
	return &rangeParametersHandlers{uc: uc}
}

// CreateRangeParameter обрабатывает POST /range.
func (h *rangeParametersHandlers) CreateRangeParameter(c *gin.Context) {
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

	created, err := h.uc.CreateRangeParameter(c.Request.Context(), &req)
	if err != nil {
		handleError(c, log, "failed to create range parameter", err)
		return
	}

	log.Info(c.Request.Context(), "action=create resource=range_parameter status=success", zap.Int64("id", created.ID))
	c.JSON(http.StatusCreated, created)
}

// ListRangeParameters обрабатывает GET /range с необязательным фильтром field_name.
func (h *rangeParametersHandlers) ListRangeParameters(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	parameters, err := h.uc.ListRangeParameters(c.Request.Context(), c.Query("field_name"))
	if err != nil {
		handleError(c, log, "failed to get range parameters", err)
		return
	}

	c.JSON(http.StatusOK, nonNil(parameters))
}

// UpdateRangeParameter обрабатывает PATCH /range/{id}.
func (h *rangeParametersHandlers) UpdateRangeParameter(c *gin.Context) {
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

	updated, err := h.uc.UpdateRangeParameter(c.Request.Context(), id, updates)
	if err != nil {
		handleError(c, log, "failed to update range parameter", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=range_parameter status=success", zap.Int64("id", id))
	c.JSON(http.StatusOK, updated)
}

// DeleteRangeParameter обрабатывает DELETE /range/{id}.
func (h *rangeParametersHandlers) DeleteRangeParameter(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	if _, err = h.uc.DeleteRangeParameter(c.Request.Context(), id); err != nil {
		handleError(c, log, "failed to delete range parameter", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=range_parameter status=success", zap.Int64("id", id))
	c.Status(http.StatusNoContent)
}
