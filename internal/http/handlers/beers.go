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

type BeersHandlers interface {
	CreateBeer(c *gin.Context)
	GetBeerById(c *gin.Context)
	UpdateBeer(c *gin.Context)
	DeleteBeer(c *gin.Context)
	GetAllBeers(c *gin.Context)
}

type beersHandler struct {
	uc usecase.BeerService
}

func NewBeersHandlers(useCase usecase.BeerService) BeersHandlers {
	return &beersHandler{
		uc: useCase,
	}
}

func (h *beersHandler) CreateBeer(c *gin.Context) {
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

func (h *beersHandler) GetBeerById(c *gin.Context) {
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

func (h *beersHandler) UpdateBeer(c *gin.Context) {
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

func (h *beersHandler) DeleteBeer(c *gin.Context) {
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

func (h *beersHandler) GetAllBeers(c *gin.Context) {
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
