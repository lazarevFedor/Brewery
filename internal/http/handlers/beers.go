package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

// BeersHandlers определяет интерфейс для обработки HTTP-запросов, связанных с пивом.
type BeersHandlers interface {
	CreateBeer(c *gin.Context)
	UpdateBeer(c *gin.Context)
	DeleteBeer(c *gin.Context)
	GetAllBeers(c *gin.Context)
	CreateBeerReview(c *gin.Context)
}

// beersHandlers реализует интерфейс BeersHandlers и использует сервис BeerService для обработки бизнес-логики.
type beersHandlers struct {
	uc usecase.BeerService
}

// NewBeersHandlers создает новый экзмепляр beersHandler с предоставленным сервисом BeerService.
func NewBeersHandlers(useCase usecase.BeerService) BeersHandlers {
	return &beersHandlers{
		uc: useCase,
	}
}

// CreateBeer обрабатывает HTTP-запрос на создание пива
func (h *beersHandlers) CreateBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Debug(c.Request.Context(), "Working")

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read beer in the request body: %v", err))

		return
	}

	var req entities.Beer
	if err = easyjson.Unmarshal(body, &req); err != nil {
		log.Error(c.Request.Context(), "failed to Unmurshal JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})

		return
	}

	beerID, err := h.uc.CreateBeer(c.Request.Context(), &req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create beer: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Debug(c.Request.Context(), fmt.Sprintf("beer_id = %d", beerID))
	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=beer status=success name=%q", req.Name))
	c.Status(http.StatusCreated)
}

// UpdateBeer обрабатывает HTTP-запрос на обновление пива.
func (h *beersHandlers) UpdateBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getIdParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid beer id: %v", err))

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

	beerID, err := h.uc.UpdateBeer(c.Request.Context(), id, updates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update beer: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Debug(c.Request.Context(), fmt.Sprintf("beer_id = %d", beerID))
	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=beer status=success id=%d", id))
	c.Status(http.StatusOK)
}

// DeleteBeer обрабатывает HTTP-запрос на удаление пива.
func (h *beersHandlers) DeleteBeer(c *gin.Context) {
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

	err = h.uc.DeleteBeer(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete beer: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=beer status=success id=%d", id))
	c.Status(http.StatusOK)
}

// GetAllBeers обрабатывает HTTP-запрос на получение всех видов пива.
func (h *beersHandlers) GetAllBeers(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid pagination params: %v", err))

		return
	}

	beers, err := h.uc.GetAllBeers(c.Request.Context(), limit, offset)
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
	log.Info(c.Request.Context(), fmt.Sprintf("action=list resource=beer status=success offset=%d limit=%d items=%d", offset, limit, len(beers)))
}

// CreateBeerReview обрабатывает HTTP-запрос на создание отзыва о пиве.
func (h *beersHandlers) CreateBeerReview(c *gin.Context) {
	reqCtx := c.Request.Context()
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	beerID, err := getBeerIDParam(c)
	if err != nil {
		log.Error(reqCtx, fmt.Sprintf("Invalid beer id: %v", err))

		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(reqCtx, fmt.Sprintf("Failed to read review in the request body: %v", err))

		return
	}

	var reviewReq entities.Review
	if err = easyjson.Unmarshal(body, &reviewReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})

		return
	}

	reviewReq.BeerID = beerID

	reviewID, err := h.uc.CreateBeerReview(c.Request.Context(), &reviewReq)
	if err != nil {
		log.Error(reqCtx, fmt.Sprintf("Failed to create review: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Debug(reqCtx, fmt.Sprintf("review_id = %d", reviewID))
	log.Info(reqCtx, fmt.Sprintf("action=create resource=review status=success beer_id=%d", beerID))
	c.Status(http.StatusOK)
}
