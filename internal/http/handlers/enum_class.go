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

// EnumService определяет интерфейс для работы с перечислениями и их значениями.
type EnumService interface {

	// CreateEnum создает новое перечисление и возвращает его ID.
	CreateEnum(ctx context.Context, enum entities.EnumClass) (uint, error)

	// GetEnum получает список перечислений для заданного поля.
	GetEnum(ctx context.Context, fieldName string) ([]entities.EnumClass, error)

	// UpdateEnum обновляет существующее перечисление по его ID.
	UpdateEnum(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnum удаляет перечисление по его ID.
	DeleteEnum(ctx context.Context, id uint) error

	// CreateEnumValue создает новое значение перечисления и возвращает его ID.
	CreateEnumValue(ctx context.Context, enum entities.EnumValue) (uint, error)

	// GetEnumValue получает список значений перечисления для заданного поля и типа.
	GetEnumValue(ctx context.Context, field string, valueType entities.EnumType) ([]entities.EnumValue, error)

	// UpdateEnumValue обновляет существующее значение перечисления по его ID.
	UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumValue удаляет значение перечисления по его ID.
	DeleteEnumValue(ctx context.Context, id uint) error
}

// EnumHandlers определяет HTTP-обработчики для классов перечислений и их значений.
type EnumHandlers interface {

	// CreateEnum обрабатывает POST /enums.
	CreateEnum(c *gin.Context)

	// GetEnum обрабатывает GET /enums.
	GetEnum(c *gin.Context)

	// UpdateEnum обрабатывает PATCH /enums/{id}.
	UpdateEnum(c *gin.Context)

	// DeleteEnum обрабатывает DELETE /enums/{id}.
	DeleteEnum(c *gin.Context)

	// CreateValue обрабатывает POST /enums/value.
	CreateValue(c *gin.Context)

	// GetValue обрабатывает GET /enums/value.
	GetValue(c *gin.Context)

	// UpdateValue обрабатывает PATCH /enums/value/{id}.
	UpdateValue(c *gin.Context)

	// DeleteValue обрабатывает DELETE /enums/value/{id}.
	DeleteValue(c *gin.Context)
}

// enumHandlers реализует интерфейс EnumHandlers.
type enumHandlers struct {
	uc EnumService
}

// NewEnumHandlers создает новый экземпляр EnumHandlers.
func NewEnumHandlers(uc EnumService) EnumHandlers {
	return &enumHandlers{uc: uc}
}

// CreateEnum обрабатывает POST /enums.
func (h *enumHandlers) CreateEnum(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req entities.EnumClass
	if !bindJSON(c, &req) {
		return
	}

	if req.FieldName == "" || req.Type == "" {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "field_name and enum_type are required")
		return
	}

	id, err := h.uc.CreateEnum(c.Request.Context(), req)
	if err != nil {
		handleError(c, log, "failed to create enum class", err)
		return
	}

	req.ID = int(id)
	log.Info(c.Request.Context(), "action=create resource=enum_class status=success", zap.Uint("id", id))
	c.JSON(http.StatusCreated, req)
}

// GetEnum обрабатывает GET /enums.
func (h *enumHandlers) GetEnum(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	fieldName := c.Query("field_name")
	if fieldName == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("field_name is required"))
		return
	}

	enums, err := h.uc.GetEnum(c.Request.Context(), fieldName)
	if err != nil {
		handleError(c, log, "failed to get enum classes", err)
		return
	}

	c.JSON(http.StatusOK, nonNil(enums))
}

// UpdateEnum обрабатывает PATCH /enums/{id}.
func (h *enumHandlers) UpdateEnum(c *gin.Context) {
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

	if err = h.uc.UpdateEnum(c.Request.Context(), toUint(id), updates); err != nil {
		handleError(c, log, "failed to update enum class", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=enum_class status=success", zap.Int64("id", id))
	c.JSON(http.StatusOK, withID(updates, id))
}

// DeleteEnum обрабатывает DELETE /enums/{id}.
func (h *enumHandlers) DeleteEnum(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	if err = h.uc.DeleteEnum(c.Request.Context(), toUint(id)); err != nil {
		handleError(c, log, "failed to delete enum class", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=enum_class status=success", zap.Int64("id", id))
	c.Status(http.StatusNoContent)
}

// CreateValue обрабатывает POST /enums/value.
func (h *enumHandlers) CreateValue(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req entities.EnumValue
	if !bindJSON(c, &req) {
		return
	}

	if req.EnumClassID <= 0 || req.ValueType == "" || req.Value == nil {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "class_id, enum_type and value are required")
		return
	}

	id, err := h.uc.CreateEnumValue(c.Request.Context(), req)
	if err != nil {
		handleError(c, log, "failed to create enum value", err)
		return
	}

	req.ID = int(id)
	log.Info(c.Request.Context(), "action=create resource=enum_value status=success", zap.Uint("id", id))
	c.JSON(http.StatusCreated, req)
}

// GetValue обрабатывает GET /enums/value.
func (h *enumHandlers) GetValue(c *gin.Context) {
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

	values, err := h.uc.GetEnumValue(c.Request.Context(), fieldName, entities.EnumType(enumType))
	if err != nil {
		handleError(c, log, "failed to get enum values", err)
		return
	}

	c.JSON(http.StatusOK, nonNil(values))
}

// UpdateValue обрабатывает PATCH /enums/value/{id}.
func (h *enumHandlers) UpdateValue(c *gin.Context) {
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

// DeleteValue обрабатывает DELETE /enums/value/{id}.
func (h *enumHandlers) DeleteValue(c *gin.Context) {
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
