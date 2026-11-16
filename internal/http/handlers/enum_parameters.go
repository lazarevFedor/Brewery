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

// EnumParametersService определяет операции с перечисляемыми параметрами.
type EnumParametersService interface {
	// CreateEnumParameter создает новый перечисляемый параметр и возвращает его ID.
	CreateEnumParameter(ctx context.Context, parameter entities.EnumClass) (uint, error)

	// ListEnumParameters получает перечисляемые параметры для заданного поля.
	ListEnumParameters(ctx context.Context, fieldName string) ([]entities.EnumClass, error)

	// UpdateEnumParameter обновляет перечисляемый параметр по его ID.
	UpdateEnumParameter(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumParameter удаляет перечисляемый параметр по его ID.
	DeleteEnumParameter(ctx context.Context, id uint) error
}

// EnumParametersHandlers определяет HTTP-обработчики перечисляемых параметров.
type EnumParametersHandlers interface {
	// CreateEnumParameter обрабатывает POST /enums.
	CreateEnumParameter(c *gin.Context)

	// ListEnumParameters обрабатывает GET /enums.
	ListEnumParameters(c *gin.Context)

	// UpdateEnumParameter обрабатывает PATCH /enums/{id}.
	UpdateEnumParameter(c *gin.Context)

	// DeleteEnumParameter обрабатывает DELETE /enums/{id}.
	DeleteEnumParameter(c *gin.Context)
}

type enumParametersHandlers struct{ uc EnumParametersService }

// NewEnumParametersHandlers создает обработчики перечисляемых параметров.
func NewEnumParametersHandlers(uc EnumParametersService) EnumParametersHandlers {
	return &enumParametersHandlers{uc: uc}
}

// CreateEnumParameter обрабатывает POST /enums.
func (h *enumParametersHandlers) CreateEnumParameter(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req enumParameter
	if !bindJSON(c, &req) {
		return
	}

	if req.FieldName == "" || req.EnumType == "" {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "field_name and enum_type are required")
		return
	}

	id, err := h.uc.CreateEnumParameter(c.Request.Context(), req.toEntity())
	if err != nil {
		handleError(c, log, "failed to create enum parameter", err)
		return
	}

	req.ID = int(id)
	log.Info(c.Request.Context(), "action=create resource=enum_parameter status=success", zap.Uint("id", id))
	c.JSON(http.StatusCreated, req)
}

// ListEnumParameters обрабатывает GET /enums.
func (h *enumParametersHandlers) ListEnumParameters(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	fieldName := c.Query("field_name")
	if fieldName == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("field_name is required"))
		return
	}

	enums, err := h.uc.ListEnumParameters(c.Request.Context(), fieldName)
	if err != nil {
		handleError(c, log, "failed to get enum parameters", err)
		return
	}

	response := make([]enumParameter, 0, len(enums))
	for _, parameter := range enums {
		response = append(response, enumParameterFromEntity(parameter))
	}
	c.JSON(http.StatusOK, response)
}

// UpdateEnumParameter обрабатывает PATCH /enums/{id}.
func (h *enumParametersHandlers) UpdateEnumParameter(c *gin.Context) {
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

	if err = h.uc.UpdateEnumParameter(c.Request.Context(), toUint(id), updates); err != nil {
		handleError(c, log, "failed to update enum parameter", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=enum_parameter status=success", zap.Int64("id", id))
	c.JSON(http.StatusOK, withID(updates, id))
}

// DeleteEnumParameter обрабатывает DELETE /enums/{id}.
func (h *enumParametersHandlers) DeleteEnumParameter(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	if err = h.uc.DeleteEnumParameter(c.Request.Context(), toUint(id)); err != nil {
		handleError(c, log, "failed to delete enum parameter", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=enum_parameter status=success", zap.Int64("id", id))
	c.Status(http.StatusNoContent)
}
