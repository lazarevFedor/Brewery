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
)

type ReviewsHandlers interface {
	CreateReview(c *gin.Context)
	UpdateReview(c *gin.Context)
	DeleteReview(c *gin.Context)
	GetBeersReviews(c *gin.Context)
}

// reviewsHandlers реализует интерфейс ReviewsHandlers и использует сервис BeerService для обработки бизнес-логики.
type reviewsHandlers struct {
	uc usecase.BeerService
}

// NewReviewsHandlers создает новый экзмепляр reviewsHandlers с предоставленным сервисом BeerService.
func NewReviewsHandlers(useCase usecase.BeerService) ReviewsHandlers {
	return &reviewsHandlers{
		uc: useCase,
	}
}

// CreateReview обрабатывает HTTP-запрос на создание отзыва о пиве.
func (h *reviewsHandlers) CreateReview(c *gin.Context) {
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

	reviewID, err := h.uc.CreateReview(c.Request.Context(), &reviewReq)
	if err != nil {
		log.Error(reqCtx, fmt.Sprintf("Failed to create review: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Debug(reqCtx, fmt.Sprintf("review_id = %d", reviewID))
	log.Info(reqCtx, fmt.Sprintf("action=create resource=review status=success beer_id=%d", beerID))
	c.Status(http.StatusOK)
}

// GetBeersReviews обрабатывает HTTP-запрос на получение всех отзывов на пиво.
func (h *reviewsHandlers) GetBeersReviews(c *gin.Context) {
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

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid pagination params: %v", err))

		return
	}

	reviews, err := h.uc.GetBeerReviews(c.Request.Context(), limit, offset, beerID)
	if err != nil {
		log.Error(reqCtx, fmt.Sprintf("Failed to get review: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	rawBytes, err := easyjson.Marshal(entities.Reviews(reviews))
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal beers: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(reqCtx, fmt.Sprintf("action=get resource=review status=success beer_id=%d", beerID))
	c.Status(http.StatusOK)
}

// UpdateReview обрабатывает HTTP-запрос на обновление отзыва на пиво.
func (h *reviewsHandlers) UpdateReview(c *gin.Context) {
	reqCtx := c.Request.Context()
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	beerID, err := getIDParam(c)
	if err != nil {
		log.Error(reqCtx, fmt.Sprintf("Invalid beer id: %v", err))

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

	err = h.uc.UpdateReview(c.Request.Context(), beerID, updates)
	if err != nil {
		log.Error(reqCtx, fmt.Sprintf("Failed to create review: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(reqCtx, fmt.Sprintf("action=update resource=review status=success beer_id=%d", beerID))
	c.Status(http.StatusOK)
}

// DeleteReviews обрабатывает HTTP-запрос на удаление на пиво.
func (h *reviewsHandlers) DeleteReview(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		c.Status(http.StatusInternalServerError)

		return
	}

	id, err := getIDParam(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))

		return
	}

	err = h.uc.DeleteReview(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete beer: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=beer status=success id=%d", id))
	c.Status(http.StatusOK)
}
