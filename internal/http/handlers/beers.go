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

// BeersHandlers определяет интерфейс для обработки HTTP-запросов, связанных с пивом.
type BeersHandlers interface {
	CreateBeer(c *gin.Context)
	UpdateBeer(c *gin.Context)
	DeleteBeer(c *gin.Context)
	GetAllBeers(c *gin.Context)
	CreateBeerReview(c *gin.Context)
}

// beersHandler реализует интерфейс BeersHandlers и использует сервис BeerService для обработки бизнес-логики.
type beersHandler struct {
	uc usecase.BeerService
}

// NewBeersHandlers создает новый экзмепляр beersHandler с предоставленным сервисом BeerService.
func NewBeersHandlers(useCase usecase.BeerService) BeersHandlers {
	return &beersHandler{
		uc: useCase,
	}
}

// CreateBeer обрабатывает HTTP-запрос на создание пива
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

	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=beer status=success name=%q", req.Name))
	c.Status(http.StatusCreated)
}

// UpdateBeer обрабатывает HTTP-запрос на обновление пива.
func (h *beersHandler) UpdateBeer(c *gin.Context) {
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

	err = h.uc.UpdateBeer(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update beer: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=beer status=success id=%d", id))
	c.Status(http.StatusOK)
}

// DeleteBeer обрабатывает HTTP-запрос на удаление пива.
func (h *beersHandler) DeleteBeer(c *gin.Context) {
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
func (h *beersHandler) GetAllBeers(c *gin.Context) {
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

	beers, err := h.uc.GetAllBeers(c.Request.Context())
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get beers: %v", err))
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
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal paginated beers: %v", w.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", w.Buffer.BuildBytes())
	log.Info(c.Request.Context(), fmt.Sprintf("action=list resource=beer status=success offset=%d limit=%d items=%d total=%d", offset, limit, len(items), total))
}

func (h *beersHandler) CreateBeerReview(c *gin.Context) {
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

	err = h.uc.CreateBeerReview(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create review: %v", err))
		c.Status(http.StatusInternalServerError)

		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=review status=success beer_id=%d", id))
	c.Status(http.StatusOK)
}
