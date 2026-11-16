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

// EnumValuesService определяет операции со значениями перечисляемых параметров.
type EnumValuesService interface {
	// CreateEnumValue создает новое значение перечисления и возвращает его ID.
	CreateEnumValue(ctx context.Context, enum entities.EnumValue) (uint, error)

	// ListEnumValues получает список значений перечисления для заданного поля и типа.
	ListEnumValues(ctx context.Context, field string, valueType entities.EnumType) ([]entities.EnumValue, error)

	// UpdateEnumValue обновляет существующее значение перечисления по его ID.
	UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumValue удаляет значение перечисления по его ID.
	DeleteEnumValue(ctx context.Context, id uint) error
}

// EnumValuesHandlers определяет HTTP-обработчики значений перечислений.
type EnumValuesHandlers interface {
	// CreateEnumValue обрабатывает POST /enums/value.
	CreateEnumValue(c *gin.Context)

	// ListEnumValues обрабатывает GET /enums/value.
	ListEnumValues(c *gin.Context)

	// UpdateEnumValue обрабатывает PATCH /enums/value/{id}.
	UpdateEnumValue(c *gin.Context)

	// DeleteEnumValue обрабатывает DELETE /enums/value/{id}.
	DeleteEnumValue(c *gin.Context)
}

type enumValuesHandlers struct{ uc EnumValuesService }

// NewEnumValuesHandlers создает обработчики значений перечислений.
func NewEnumValuesHandlers(uc EnumValuesService) EnumValuesHandlers {
	return &enumValuesHandlers{uc: uc}
}

// CreateEnumValue обрабатывает POST /enums/value.
func (h *enumValuesHandlers) CreateEnumValue(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req enumValue
	if !bindJSON(c, &req) {
		return
	}

	if req.ParameterID <= 0 || req.EnumType == "" || req.Value == nil {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "parameter_id, enum_type and value are required")
		return
	}

	id, err := h.uc.CreateEnumValue(c.Request.Context(), req.toEntity())
	if err != nil {
		handleError(c, log, "failed to create enum value", err)
		return
	}

	req.ID = int(id)
	log.Info(c.Request.Context(), "action=create resource=enum_value status=success", zap.Uint("id", id))
	c.JSON(http.StatusCreated, req)
}

// ListEnumValues обрабатывает GET /enums/value.
func (h *enumValuesHandlers) ListEnumValues(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	fieldName := c.Query("field_name")
	enumType := c.Query("enum_type")
	if fieldName == "" || enumType == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("field_name and enum_type are required"))
		return
	}

	values, err := h.uc.ListEnumValues(c.Request.Context(), fieldName, entities.EnumType(enumType))
	if err != nil {
		handleError(c, log, "failed to get enum values", err)
		return
	}

	response := make([]enumValue, 0, len(values))
	for _, value := range values {
		response = append(response, enumValueFromEntity(value))
	}
	c.JSON(http.StatusOK, response)
}

// UpdateEnumValue обрабатывает PATCH /enums/value/{id}.
func (h *enumValuesHandlers) UpdateEnumValue(c *gin.Context) {
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

	if err = h.uc.UpdateEnumValue(c.Request.Context(), toUint(id), updates); err != nil {
		handleError(c, log, "failed to update enum value", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=enum_value status=success", zap.Int64("id", id))
	c.JSON(http.StatusOK, withID(updates, id))
}

// DeleteEnumValue обрабатывает DELETE /enums/value/{id}.
func (h *enumValuesHandlers) DeleteEnumValue(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	if err = h.uc.DeleteEnumValue(c.Request.Context(), toUint(id)); err != nil {
		handleError(c, log, "failed to delete enum value", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=enum_value status=success", zap.Int64("id", id))
	c.Status(http.StatusNoContent)
}

// withID добавляет id к измененным полям. Сервис перечислений после обновления не возвращает объект,
// поэтому в ответ отдаем то, что поменяли.
func withID(updates map[string]any, id int64) map[string]any {
	updates["id"] = id
	return updates
}
