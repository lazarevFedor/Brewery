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

type BeersHandlers interface {
	CreateBeer(c *gin.Context)
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

	page, limit, err := getPaginationParams(c)
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

	offset := (page - 1) * limit
	items := make([]entities.Beer, 0)
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}

		items = beers[offset:end]
	}

	var w jwriter.Writer
	w.RawByte('{')
	w.RawString("\"items\":")
	entities.Beers(items).MarshalEasyJSON(&w)
	w.RawString(",\"page\":")
	w.Int(page)
	w.RawString(",\"limit\":")
	w.Int(limit)
	w.RawString(",\"total\":")
	w.Int(total)
	w.RawString(",\"total_pages\":")
	w.Int(totalPages)
	w.RawString(",\"has_next\":")
	w.Bool(page < totalPages)
	w.RawString(",\"has_prev\":")
	w.Bool(page > 1)
	w.RawByte('}')

	if w.Error != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal paginated beers: %v", w.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})

		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", w.Buffer.BuildBytes())
	log.Info(c.Request.Context(), "")
}
