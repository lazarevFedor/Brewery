// Package handlers содержит реализацию HTTP-обработчиков для управления пивоварней.
package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"

	"Brewery/internal/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
)

// CategoriesHandlers определяет интерфейс для обработки HTTP-запросов, связанных с категориями продуктов.
type CategoriesHandlers interface {

	// CreateCategory обрабатывает HTTP-запрос на создание новой категории продукта.
	CreateCategory(c *gin.Context)

	// GetCategoryByID обрабатывает HTTP-запрос на получение категории продукта по ее идентификатору.
	GetCategoryByID(c *gin.Context)

	// UpdateCategory обрабатывает HTTP-запрос на обновление существующей категории продукта по ее идентификатору.
	UpdateCategory(c *gin.Context)

	// DeleteCategory обрабатывает HTTP-запрос на удаление категории по ее идентификатору.
	DeleteCategory(c *gin.Context)

	// GetAllCategories обрабатывает HTTP-запрос на получение всех категорий продуктов.
	GetAllCategories(c *gin.Context)

	// GetBeersByCategory обрабатывает HTTP-запрос на получение пива по заданной идентификатором категории.
	GetBeersByCategory(c *gin.Context)

	// GetParentCategory обрабатывает HTTP-запрос на получение родительской категории для заданной категории по ее идентификатору.
	GetParentCategory(c *gin.Context)

	// GetChildCategory обрабатывает HTTP-запрос на получение дочерней категории для заданной категории по ее идентификатору.
	GetChildCategory(c *gin.Context)
}

// categoriesHandlers конкретная реализация интерфейса CategoriesHandlers, которая использует сервис BeerService для обработки бизнес-логики.
type categoriesHandlers struct {
	uc usecase.BeerService
}

// NewCategoriesHandlers создает новый экземпляр categoriesHandler с предоставленным сервисом BeerService.
func NewCategoriesHandlers(useCase usecase.BeerService) CategoriesHandlers {
	return &categoriesHandlers{
		uc: useCase,
	}
}

// CreateCategory обрабатывает HTTP-запрос на создание новой категории продукта.
func (h *categoriesHandlers) CreateCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read category in the request body: %v", err))
		return
	}

	var req entities.ProductCategory

	if err = easyjson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to unmarshal category: %v", err))
		return
	}

	ctgID, err := h.uc.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get category: %v", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	log.Debug(c.Request.Context(), fmt.Sprintf("ctgID=%d", ctgID))
	log.Info(
		c.Request.Context(),
		fmt.Sprintf("action=create resource=category status=success name=%q", req.Name),
	)
	c.Status(http.StatusCreated)
}

// GetCategoryByID обрабатывает HTTP-запрос на получение категории продукта по ее идентификатору.
//
//nolint:staticcheck, ineffassign, wastedassign
func (h *categoriesHandlers) GetCategoryByID(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})

		return
	}

	categoryID, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	category, err := h.uc.GetCategoryByID(c.Request.Context(), categoryID)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get category: %v", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})

		return
	}

	rawBytes, err := easyjson.Marshal(category)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal category: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal category"})

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=get resource=category status=success id=%d", categoryID))
}

// UpdateCategory обрабатывает HTTP-запрос на обновление существующей категории продукта по ее идентификатору.
func (h *categoriesHandlers) UpdateCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})

		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

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

	err = h.uc.UpdateCategory(c.Request.Context(), id, updates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update category: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=category status=success id=%d", id))
	c.Status(http.StatusOK)
}

// DeleteCategory обрабатывает HTTP-запрос на удаление категории по ее идентификатору.
func (h *categoriesHandlers) DeleteCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})

		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	err = h.uc.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete category: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=category status=success id=%d", id))
	c.Status(http.StatusOK)
}

// GetAllCategories обрабатывает HTTP-запрос на получение всех категорий продуктов.
func (h *categoriesHandlers) GetAllCategories(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	categories, err := h.uc.GetAllCategories(c.Request.Context())
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get all categories: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get all categories"})

		return
	}

	rawBytes, err := easyjson.Marshal(entities.Products(categories))
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal categories: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal categories"})

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=list resource=category status=success items=%d", len(categories)))
}

// GetParentCategory обрабатывает HTTP-запрос на получение родительской категории для заданной категории по ее идентификатору.
func (h *categoriesHandlers) GetParentCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	category, err := h.uc.GetParentCategory(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get parent category: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	rawBytes, err := easyjson.Marshal(category)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal category: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=get resource=category_parent status=success id=%d", id))
}

// GetChildCategory обрабатывает HTTP-запрос на получение дочерней категории для заданной категории по ее идентификатору.
func (h *categoriesHandlers) GetChildCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	children, err := h.uc.GetChildCategories(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get child category: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	rawBytes, err := easyjson.Marshal(entities.Products(children))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal category: %v", err))

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=get resource=category_child status=success id=%d", id))
}

// GetBeersByCategory обрабатывает HTTP-запрос на получение пива по заданной идентификатором категории.
func (h *categoriesHandlers) GetBeersByCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	id, err := getUintParam(c, "category_id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidID, "Invalid category id")
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid pagination params: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidParameters, "Invalid pagination parameters")
		return
	}

	beers, err := h.uc.GetBeersByCategory(c.Request.Context(), id, limit, offset)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get beers by category: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	rawBytes, err := easyjson.Marshal(entities.Beers(beers))
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal beers: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=list resource=beer_by_category status=success category_id=%d offset=%d limit=%d items=%d", id, offset, limit, len(beers)))
}
