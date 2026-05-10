package handlers

import (
	"Brewery/internal/entities"
	"Brewery/internal/apperrors"
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

	GetFeature(c *gin.Context)
	CreateFeature(c *gin.Context)
	DeleteFeature(c *gin.Context)
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

func (h *beersHandlers) DeleteFeature(c *gin.Context) {
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

	err = h.uc.DeleteFeature(c.Request.Context(), id)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to delete beer's feature: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	log.Info(c.Request.Context(), fmt.Sprintf("action=delete resource=beer status=success id=%d", id))
	c.Status(http.StatusNoContent)
}
