package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
)

type CategoriesHandlers interface {
	CreateCategory(c *gin.Context)
	GetCategoryById(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
	GetAllCategories(c *gin.Context)

	GetParentCategory(c *gin.Context)
	GetChildCategory(c *gin.Context)
}

type categoriesHandler struct {
	uc usecase.BeerService
}

func NewCategoriesHandlers(useCase usecase.BeerService) CategoriesHandlers {
	return &categoriesHandler{
		uc: useCase,
	}
}

func (h *categoriesHandler) CreateCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		log.Error(c.Request.Context(), "Failed to get logger from context")

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

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusCreated)
}

func (h *categoriesHandler) GetCategoryById(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		log.Error(c.Request.Context(), "Failed to get logger from context")

		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	category, err := h.uc.GetCategoryById(c.Request.Context(), id)
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
	log.Info(c.Request.Context(), "")
}

func (h *categoriesHandler) UpdateCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		log.Error(c.Request.Context(), "Failed to get logger from context")

		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	err = h.uc.UpdateCategory(c, id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update category: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})

		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *categoriesHandler) DeleteCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		log.Error(c.Request.Context(), "Failed to get logger from context")

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

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *categoriesHandler) GetAllCategories(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		log.Error(c.Request.Context(), "Failed to get logger from context")

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
	log.Info(c.Request.Context(), "")
}

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
	log.Info(c.Request.Context(), "")
}

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
	log.Info(c.Request.Context(), "")
}
