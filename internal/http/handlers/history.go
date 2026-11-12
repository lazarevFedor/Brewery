package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HistoryService определяет бизнес-логику работы с историей изменений спецификаций.
// Сигнатуры совпадают с usecase.HistoryService.
type HistoryService interface {

	// CreateChangeRecord регистрирует факт изменения спецификации изделия (REGISTER_CHANGE).
	CreateChangeRecord(ctx context.Context, record *entities.HistoryRecord) (*entities.HistoryRecord, error)

	// GetHistoryByProduct возвращает историю изменений изделия за период. Нулевые start/end - без ограничения.
	GetHistoryByProduct(ctx context.Context, productID int64, start, end time.Time, limit, offset int32) ([]entities.HistoryRecord, error)

	// GetHistoryRecordByID возвращает запись истории по id (пустой срез, если записи нет).
	GetHistoryRecordByID(ctx context.Context, id int64) ([]entities.HistoryRecord, error)
}

// HistoryHandlers - HTTP-обработчики для истории изменений.
type HistoryHandlers interface {

	// RegisterProductChange обрабатывает POST /products/{id}/history.
	RegisterProductChange(c *gin.Context)

	// ListProductHistory обрабатывает GET /products/{id}/history.
	ListProductHistory(c *gin.Context)

	// GetHistoryRecord обрабатывает GET /history/{id}.
	GetHistoryRecord(c *gin.Context)
}

// historyRecordResponse - представление записи истории (схема HistoryRecord в swagger).
type historyRecordResponse struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	Changes   string    `json:"changes"`
	CreatedAt time.Time `json:"created_at"`
}

func newHistoryRecordResponse(r entities.HistoryRecord) historyRecordResponse {
	return historyRecordResponse(r)
}

// historyRecordCreateRequest - тело POST /products/{id}/history (схема HistoryRecordCreate).
type historyRecordCreateRequest struct {
	Changes string `json:"changes"`
}

// historyHandlers - реализация HistoryHandlers.
type historyHandlers struct {
	uc HistoryService
}

// NewHistoryHandlers создает обработчики истории изменений. Если uc == nil, все запросы отвечают 501.
func NewHistoryHandlers(uc HistoryService) HistoryHandlers {
	return &historyHandlers{uc: uc}
}

// RegisterProductChange обрабатывает POST /products/{id}/history.
func (h *historyHandlers) RegisterProductChange(c *gin.Context) {
	if h.uc == nil {
		notImplemented(c)
		return
	}

	log, ok := getLogger(c)
	if !ok {
		return
	}

	productID, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	var req historyRecordCreateRequest
	if !bindJSON(c, &req) {
		return
	}

	req.Changes = strings.TrimSpace(req.Changes)
	if req.Changes == "" {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "changes is required")
		return
	}

	record, err := h.uc.CreateChangeRecord(c.Request.Context(), &entities.HistoryRecord{
		ProductID: productID,
		Changes:   req.Changes,
	})
	if err != nil {
		handleError(c, log, "failed to register change", err)
		return
	}

	log.Info(c.Request.Context(), "action=create resource=history_record status=success",
		zap.Int64("product_id", productID), zap.Int64("id", record.ID))
	c.JSON(http.StatusCreated, newHistoryRecordResponse(*record))
}

// ListProductHistory обрабатывает GET /products/{id}/history?from=&to=.
func (h *historyHandlers) ListProductHistory(c *gin.Context) {
	if h.uc == nil {
		notImplemented(c)
		return
	}

	log, ok := getLogger(c)
	if !ok {
		return
	}

	productID, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	from, err := getTimeQuery(c, "from")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	to, err := getTimeQuery(c, "to")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	if !from.IsZero() && !to.IsZero() && from.After(to) {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("from must not be after to"))
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	records, err := h.uc.GetHistoryByProduct(c.Request.Context(), productID, from, to, limit, offset)
	if err != nil {
		handleError(c, log, "failed to list product history", err)
		return
	}

	resp := make([]historyRecordResponse, 0, len(records))
	for _, r := range records {
		resp = append(resp, newHistoryRecordResponse(r))
	}

	c.JSON(http.StatusOK, resp)
}

// GetHistoryRecord обрабатывает GET /history/{id}.
func (h *historyHandlers) GetHistoryRecord(c *gin.Context) {
	if h.uc == nil {
		notImplemented(c)
		return
	}

	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	records, err := h.uc.GetHistoryRecordByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, log, "failed to get history record", err)
		return
	}

	if len(records) == 0 {
		writeError(c, http.StatusNotFound, apperrors.CodeNotFound, "History record not found")
		return
	}

	c.JSON(http.StatusOK, newHistoryRecordResponse(records[0]))
}
