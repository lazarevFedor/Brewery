package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ClassService определяет бизнес-логику работы с классами изделий, которая нужна HTTP-слою.
type ClassService interface {
	GetAllClasses(ctx context.Context) ([]entities.Class, error)
	CreateClass(ctx context.Context, class *entities.Class) (*entities.Class, error)
	GetClassByID(ctx context.Context, id int64) (*entities.Class, error)
	UpdateClass(ctx context.Context, id int64, updates map[string]any) (*entities.Class, error)
	DeleteClassByID(ctx context.Context, id int64) error
	GetParentClass(ctx context.Context, id int64) (*entities.Class, error)
	GetChildClasses(ctx context.Context, id int64) ([]entities.Class, error)
}

// ClassHandlers определяет HTTP-обработчики для классов изделий.
type ClassHandlers interface {

	// CreateClass обрабатывает POST /classes.
	CreateClass(c *gin.Context)

	// GetClass обрабатывает GET /classes/{id}.
	GetClass(c *gin.Context)

	// UpdateClass обрабатывает PATCH /classes/{id}.
	UpdateClass(c *gin.Context)

	// DeleteClass обрабатывает DELETE /classes/{id}.
	DeleteClass(c *gin.Context)

	// ListClasses обрабатывает GET /classes.
	ListClasses(c *gin.Context)

	// GetClassParent обрабатывает GET /classes/parent/{id}.
	GetClassParent(c *gin.Context)

	// ListClassChildren обрабатывает GET /classes/children/{id}.
	ListClassChildren(c *gin.Context)
}

// classHandlers реализует ClassHandlers поверх ClassService.
type classHandlers struct {
	uc ClassService
}

// NewClassHandlers создает обработчики классов изделий.
func NewClassHandlers(uc ClassService) ClassHandlers {
	return &classHandlers{uc: uc}
}

// CreateClass обрабатывает POST /classes.
func (h *classHandlers) CreateClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req entities.Class
	if !bindJSON(c, &req) {
		return
	}

	if req.Name == "" {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "Class name is required")
		return
	}

	created, err := h.uc.CreateClass(c.Request.Context(), &req)
	if err != nil {
		handleError(c, log, "failed to create class", err)
		return
	}

	log.Info(c.Request.Context(), "action=create resource=class status=success", zap.Int64("id", created.ID))
	c.JSON(http.StatusCreated, created)
}

// GetClass обрабатывает GET /classes/{id}.
func (h *classHandlers) GetClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	class, err := h.uc.GetClassByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, log, "failed to get class", err)
		return
	}

	if class == nil {
		writeError(c, http.StatusNotFound, apperrors.CodeNotFound, "Class not found")
		return
	}

	c.JSON(http.StatusOK, class)
}

// UpdateClass обрабатывает PATCH /classes/{id}.
func (h *classHandlers) UpdateClass(c *gin.Context) {
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

	updated, err := h.uc.UpdateClass(c.Request.Context(), id, updates)
	if err != nil {
		handleError(c, log, "failed to update class", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=class status=success", zap.Int64("id", id))
	c.JSON(http.StatusOK, updated)
}

// DeleteClass обрабатывает DELETE /classes/{id}.
func (h *classHandlers) DeleteClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	if err = h.uc.DeleteClassByID(c.Request.Context(), id); err != nil {
		handleError(c, log, "failed to delete class", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=class status=success", zap.Int64("id", id))
	c.Status(http.StatusNoContent)
}

// ListClasses обрабатывает GET /classes.
func (h *classHandlers) ListClasses(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classes, err := h.uc.GetAllClasses(c.Request.Context())
	if err != nil {
		handleError(c, log, "failed to list classes", err)
		return
	}

	c.JSON(http.StatusOK, nonNil(classes))
}

// GetClassParent обрабатывает GET /classes/parent/{id}.
func (h *classHandlers) GetClassParent(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	parent, err := h.uc.GetParentClass(c.Request.Context(), id)
	if err != nil {
		handleError(c, log, "failed to get parent class", err)
		return
	}

	// У корневого класса родителя нет.
	if parent == nil {
		writeError(c, http.StatusNotFound, apperrors.CodeNotFound, "Parent class not found")
		return
	}

	c.JSON(http.StatusOK, parent)
}

// ListClassChildren обрабатывает GET /classes/children/{id}.
func (h *classHandlers) ListClassChildren(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	children, err := h.uc.GetChildClasses(c.Request.Context(), id)
	if err != nil {
		handleError(c, log, "failed to get child classes", err)
		return
	}

	c.JSON(http.StatusOK, nonNil(children))
}

// nonNil возвращает пустой срез вместо nil, чтобы в JSON был [] а не null.
func nonNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}

	return items
}
