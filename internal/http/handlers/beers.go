// Package handlers содержит реализацию HTTP-обработчиков для управления пивоварней.
package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

// BeersHandlers определяет интерфейс для обработки HTTP-запросов, связанных с пивом.
type BeersHandlers interface {

	// CreateBeer обрабатывает HTTP-запрос на создание пива.
	CreateBeer(c *gin.Context)

	// UpdateBeer обрабатывает HTTP-запрос на обновление пива.
	UpdateBeer(c *gin.Context)

	// DeleteBeer обрабатывает HTTP-запрос на удаление пива.
	DeleteBeer(c *gin.Context)

	// GetAllBeers обрабатывает HTTP-запрос на получение всех видов пива.
	GetAllBeers(c *gin.Context)

	// SearchBeer обрабатывает HTTP-запрос на поиск пива по заданным фильтрам.
	SearchBeer(c *gin.Context)

	// GetFeature обрабатывает HTTP-запрос на получение характеристик пива.
	GetFeature(c *gin.Context)

	// CreateFeature обрабатывает HTTP-запрос на создание характеристики пива.
	CreateFeature(c *gin.Context)

	// DeleteFeature обрабатывает HTTP-запрос на удаление характеристики пива.
	DeleteFeature(c *gin.Context)

	GetBeerByID(c *gin.Context)
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

func (h *beersHandlers) GetBeerByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	beer, err := h.uc.GetBeerByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "beer not found"})
		return
	}
	c.JSON(http.StatusOK, beer)
}

// CreateBeer обрабатывает HTTP-запрос на создание пива.
func (h *beersHandlers) CreateBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read request body: %v", err))
		handleError(c, err)
		return
	}

	var req entities.Beer
	if err = easyjson.Unmarshal(body, &req); err != nil {
		log.Error(c.Request.Context(), "failed to Unmurshal JSON", zap.Error(err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidJSON, "Request body is not valid JSON")
		return
	}

	beer, err := h.uc.CreateBeer(c.Request.Context(), &req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create beer: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")

		return
	}

	log.Debug(c.Request.Context(), fmt.Sprintf("beer_id = %d", beer.ID))
	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=beer status=success name=%q", req.Name))
	c.JSON(http.StatusCreated, beer)
}

// UpdateBeer обрабатывает HTTP-запрос на обновление пива.
func (h *beersHandlers) UpdateBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid beer id: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidID, "Invalid beer id")
		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read request body: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeBadRequest, "Failed to read request body")
		return
	}

	updates := make(map[string]any)
	if err = json.Unmarshal(body, &updates); err != nil {
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidJSON, "Request body is not valid JSON")
		return
	}

	if len(updates) == 0 {
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidJSON, "Request body is empty")
		return
	}

	beer, err := h.uc.UpdateBeer(c.Request.Context(), id, updates)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to update beer: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	log.Debug(c.Request.Context(), fmt.Sprintf("beer_id = %d", beer.ID))
	log.Info(c.Request.Context(), fmt.Sprintf("action=update resource=beer status=success id=%d", id))
	c.JSON(http.StatusOK, beer)
}

// DeleteBeer обрабатывает HTTP-запрос на удаление пива.
func (h *beersHandlers) DeleteBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid beer id: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidID, "Invalid beer id")
		return
	}

	err = h.uc.DeleteBeer(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete beer: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=beer status=success id=%d", id))
	c.Status(http.StatusNoContent)
}

// GetAllBeers обрабатывает HTTP-запрос на получение всех видов пива.
func (h *beersHandlers) GetAllBeers(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid pagination params: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidParameters, "Invalid pagination parameters")
		return
	}

	beers, err := h.uc.GetAllBeers(c.Request.Context(), limit, offset)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get beers: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")

		return
	}

	rawBytes, err := easyjson.Marshal(entities.Beers(beers))
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to marshal beers: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", rawBytes)
	log.Info(c.Request.Context(), fmt.Sprintf(
		"action=list resource=beer status=success offset=%d limit=%d items=%d",
		offset, limit, len(beers),
	))
}

// SearchBeer обрабатывает HTTP-запрос на поиск пива по заданным фильтрам.
func (h *beersHandlers) SearchBeer(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}
	log.Debug(c.Request.Context(), "after logger")

	categoryID, err := getUintParam(c, "id")
	if err != nil {
		if err.Error() == "invalid id" {
			log.Debug(c.Request.Context(), "get uint param", zap.Int("id", int(categoryID)))
			log.Error(c.Request.Context(), fmt.Sprintf("Invalid category id: %v", err))
			writeError(c, http.StatusBadRequest, apperrors.CodeInvalidID, "Invalid category id")
			return
		}
	}

	filters := strings.Split(c.Query("filter"), "&")
	if len(filters) == 0 {
		log.Error(c.Request.Context(), "Missing filter parameters")
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidParameters, "Missing filter parameters")
		return
	}

	validatedFilters := make([]*entities.FilterParameter, len(filters))
	for i, filter := range filters {
		vf, err := validateFilterParam(filter)
		if err != nil {
			log.Error(c.Request.Context(), fmt.Sprintf("Invalid filter parameter: %v", err))
			writeError(c, http.StatusBadRequest, apperrors.CodeInvalidParameters, "Invalid filter parameters")
			return
		}
		validatedFilters[i] = vf
	}

	log.Debug(c.Request.Context(), "validated filters", zap.Any("filters", validatedFilters))
	offset, limit, err := getPaginationParams(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid pagination params: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidParameters, "Invalid pagination parameters")
		return
	}

	filteredBeers, err := h.uc.FilterBeer(c.Request.Context(), validatedFilters, limit, offset, categoryID)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to filter beers: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInvalidParameters, "Unexpected error occurred")
		return
	}

	log.Info(
		c.Request.Context(),
		fmt.Sprintf(
			"action=list resource=beer status=success offset=%d limit=%d items=%d",
			offset, limit, len(filteredBeers),
		),
	)
	c.JSON(http.StatusOK, filteredBeers)
}

// GetFeature обрабатывает HTTP-запрос на получение характеристик пива.
func (h *beersHandlers) GetFeature(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid beer id: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidID, "Invalid beer id")
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid pagination params: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidParameters, "Invalid pagination parameters")
		return
	}

	feats, err := h.uc.GetFeatures(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to get beer's features: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	log.Info(
		c.Request.Context(),
		fmt.Sprintf(
			"action=list resource=beer status=success offset=%d limit=%d items=%d",
			offset, limit, len(feats),
		),
	)
	c.JSON(http.StatusOK, feats)
}

// CreateFeature обрабатывает HTTP-запрос на создание характеристики пива.
func (h *beersHandlers) CreateFeature(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	body, err := readRequestBody(c)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to read request body: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeBadRequest, "Failed to read request body")
		return
	}

	id, err := getUintParam(c, "id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid beer id: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidID, "Invalid beer id")
		return
	}

	var featName string
	if err = json.Unmarshal(body, &featName); err != nil {
		log.Error(c.Request.Context(), "failed to Unmurshal JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	featID, err := h.uc.CreateFeature(c.Request.Context(), id, featName)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create beer's feature: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=beer status=success name=%q", featName))
	c.JSON(http.StatusCreated, gin.H{
		"id":   featID,
		"name": featName,
	})
}

// DeleteFeature обрабатывает HTTP-запрос на удаление характеристики пива.
func (h *beersHandlers) DeleteFeature(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	id, err := getUintParam(c, "beer_id")
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Invalid beer id: %v", err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidID, "Invalid beer id")
		return
	}

	err = h.uc.DeleteFeature(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete beer's feature: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=beer status=success id=%d", id))
	c.Status(http.StatusNoContent)
}
