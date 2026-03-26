// Package handlers содержит реализацию HTTP-обработчиков для управления категориями продуктов в пивоварне.
package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jwriter"
)

// CategoriesHandlers определяет интерфейс для обработки HTTP-запросов, связанных с категориями продуктов.
type CategoriesHandlers interface {
	CreateCategory(c *gin.Context)
	GetCategoryById(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
	GetAllCategories(c *gin.Context)

	GetBeersByCategory(c *gin.Context)
	GetParentCategory(c *gin.Context)
	GetChildCategory(c *gin.Context)
}

// categoriesHandler конкретная реализация интерфейса CategoriesHandlers, которая использует сервис BeerService для обработки бизнес-логики.
type categoriesHandler struct {
	uc usecase.BeerService
}

// NewCategoriesHandlers создает новый экземпляр categoriesHandler с предоставленным сервисом BeerService.
func NewCategoriesHandlers(useCase usecase.BeerService) CategoriesHandlers {
	return &categoriesHandler{
		uc: useCase,
	}
}

// CreateCategory обрабатывает HTTP-запрос на создание новой категории продукта.
func (h *categoriesHandler) CreateCategory(c *gin.Context) {
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

	log.Info(
		c.Request.Context(),
		// fmt.Sprintf("action=create resource=category status=success name=%q level=%d",
		// 	req.Name, req.Level))
		fmt.Sprintf("action=create resource=category status=success name=%q",
			req.Name))
	c.Status(http.StatusCreated)
}

// GetCategoryById обрабатывает HTTP-запрос на получение категории продукта по ее идентификатору.
func (h *categoriesHandler) GetCategoryById(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})

		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	

	category, err := h.uc.GetCategoryByID(c.Request.Context(), id)
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
	log.Info(c.Request.Context(), fmt.Sprintf("action=get resource=category status=success id=%d", id))
	}
}

// UpdateCategory обрабатывает HTTP-запрос на обновление существующей категории продукта по ее идентификатору.
func (h *categoriesHandler) UpdateCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	err = h.uc.UpdateCategory(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update category: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=category status=success id=%d", id))
	c.Status(http.StatusOK)
}

// DeleteCategory обрабатывает HTTP-запрос на удаление категории по ее идентификатору.
func (h *categoriesHandler) DeleteCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})

		return
	}

	id, err := getIdParam(c)
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
func (h *categoriesHandler) GetAllCategories(c *gin.Context) {
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
func (h *categoriesHandler) GetParentCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getIdParam(c)
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
func (h *categoriesHandler) GetChildCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	category, err := h.uc.GetChildCategory(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get child category: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	rawBytes, err := easyjson.Marshal(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal category: %v", err))

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf("action=get resource=category_child status=success id=%d", id))
}

// GetBeersByCategory обрабатывает HTTP-запрос на получение пива по заданной идентификатором категории.
func (h *categoriesHandler) GetBeersByCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid pagination params: %v", err))

		return
	}

	beers, err := h.uc.GetBeersByCategory(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get beers by category: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	total := len(beers)

	totalPages := 0

	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	items := make([]entities.Beer, 0)

	if offset < total {
		end := min(offset+limit, total)

		items = beers[offset:end]
	}

	var w jwriter.Writer
	w.RawByte('{')
	w.RawString("\"items\":")
	entities.Beers(items).MarshalEasyJSON(&w)
	w.RawString(",\"offset\":")
	w.Int(offset)
	w.RawString(",\"limit\":")
	w.Int(limit)
	w.RawString(",\"total\":")
	w.Int(total)
	w.RawString(",\"total_pages\":")
	w.Int(totalPages)
	w.RawString(",\"has_next\":")
	w.Bool(offset+limit < total)
	w.RawString(",\"has_prev\":")
	w.Bool(offset > 0)
	w.RawByte('}')

	if w.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal beer: %v", w.Error))

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", w.Buffer.BuildBytes())
	log.Info(c.Request.Context(), fmt.Sprintf("action=list resource=beer_by_category status=success category_id=%d offset=%d limit=%d items=%d total=%d", id, offset, limit, len(items), total))
}
