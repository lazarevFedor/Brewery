package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
)

type BreweryHandlers interface {
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
	GetAllBeers(c *gin.Context)
}

type BeersHandler struct {
	uc usecase.BeerService
}

func NewBreweryHandlers(useCase usecase.BeerService) BreweryHandlers {
	return &BeersHandler{
		uc: useCase,
	}
}

func getIdParam(c *gin.Context) (int, error) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})

		return 0, errors.New("invalid category id")
	}

	return id, nil
}

func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})

		return nil, errors.New("invalid request body")
	}

	return body, nil
}

func (h *BeersHandler) CreateCategory(c *gin.Context) {
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

func (h *BeersHandler) GetCategoryById(c *gin.Context) {
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

func (h *BeersHandler) UpdateCategory(c *gin.Context) {
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

func (h *BeersHandler) DeleteCategory(c *gin.Context) {
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

func (h *BeersHandler) GetAllCategories(c *gin.Context) {
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

func (h *BeersHandler) GetParentCategory(c *gin.Context) {
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

func (h *BeersHandler) GetChildCategory(c *gin.Context) {
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

func (h *BeersHandler) CreateBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read beer in the request body: %v", err))

		return
	}

	var req entities.Beer
	if err = easyjson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})

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

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	beer, err := h.uc.GetBeerById(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get beer by id: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	rawBytes, err := easyjson.Marshal(beer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal beer: %v", err))

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), "")
}

func (h *BeersHandler) UpdateBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)
		log.Error(c.Request.Context(), "Failed to get logger from context")

		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	err = h.uc.UpdateBeer(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update beer: %v", err))
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
		log.Error(c.Request.Context(), "Failed to get logger from context")

		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	err = h.uc.DeleteBeer(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete beer: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), "")
	c.Status(http.StatusOK)
}

func (h *BeersHandler) GetAllBeers(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	beers, err := h.uc.GetAllBeers(c.Request.Context())
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get beers: %v", err))
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
	log.Info(c.Request.Context(), "")
}
