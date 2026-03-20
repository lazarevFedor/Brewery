package handlers

import (
	"Brewery/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type beerClassifier interface {
	CreateCategory(c *gin.Context)
	GetCategoryById(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
	GetAllCategories(c *gin.Context)

	GetParentCategory(c *gin.Context)
	GetChildCategory(c *gin.Context)

	CreateBeer(c *gin.Context)
	GetBeerById(c *gin.Context)
	UpdateBeer(c *gin.Context)
	DeleteBeer(c *gin.Context)
}

type BeersHandler struct {
	beerCl beerClassifier
}

func NewBeersHandler(bc beerClassifier) *BeersHandler {
	return &BeersHandler{
		beerCl: bc,
	}
}

func (h *BeersHandler) CreateCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusCreated)
}

func (h *BeersHandler) GetCategoryById(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) UpdateCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) DeleteCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logger from context"})
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) GetAllCategories(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) GetParentCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) GetChildCategory(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) CreateBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusCreated)
}

func (h *BeersHandler) GetBeerById(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) UpdateBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) DeleteBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}
