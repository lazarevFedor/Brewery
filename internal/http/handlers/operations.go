package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// OperationService определяет интерфейс сервисного слоя для работы с технологическими операциями.
type OperationService interface {
	// CreateOperation создает операцию.
	CreateOperation(ctx context.Context, op entities.Operation) (*entities.Operation, error)

	// GetOperations возвращает список операций с учетом пагинации и опциональной фильтрации по материалу и продукту.
	GetOperations(ctx context.Context, limit, offset int32, materialID, productID *int64) ([]entities.Operation, error)

	// GetOperationByID возвращает операцию по ID с заполненными связями (оборудование, материалы).
	GetOperationByID(ctx context.Context, id int64) (*entities.Operation, error)

	// UpdateOperation обновляет операцию.
	UpdateOperation(ctx context.Context, id int64, updates map[string]any) (*entities.Operation, error)

	// DeleteOperation удаляет операцию по идентификатору.
	DeleteOperation(ctx context.Context, id int64) error

	// GetProductTechCard возвращает операции, нормы расхода и суммарную стоимость производства.
	GetProductTechCard(ctx context.Context, productID int64, quantity float32) (*entities.ProductTechCard, error)
}

// OperationHandlers определяет HTTP-обработчики для операций.
type OperationHandlers interface {
	// ListOperations обрабатывает GET /operations.
	ListOperations(c *gin.Context)

	// CreateOperation обрабатывает POST /operations.
	CreateOperation(c *gin.Context)

	// GetOperation обрабатывает GET /operations/{id}.
	GetOperation(c *gin.Context)

	// UpdateOperation обрабатывает PATCH /operations/{id}.
	UpdateOperation(c *gin.Context)

	// DeleteOperation обрабатывает DELETE /operations/{id}.
	DeleteOperation(c *gin.Context)

	// GetProductTechCard обрабатывает GET /products/{id}/operations.
	GetProductTechCard(c *gin.Context)
}

// productShortResponse - краткое представление продукта.
type productShortResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	IsFinalProduct bool   `json:"is_final_product,omitempty"`
}

// equipmentShortResponse - краткое представление оборудования.
type equipmentShortResponse struct {
	Name       string `json:"name"`
	WorkCenter string `json:"work_center"`
}

// equipmentRequest - представление оборудования в запросе операции.
type equipmentRequest struct {
	Name       string `json:"name"`
	WorkCenter string `json:"work_center"`
}

// operationResponse - представление операции в API.
type operationResponse struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ClassID      int64                  `json:"operation_class_id"`
	ProfessionID int64                  `json:"profession_id"`
	StepNumber   int32                  `json:"step_number,omitempty"`
	LaborCost    float64                `json:"labor_cost"`
	Parameters   map[string]any         `json:"parameters,omitempty"`
	Equipment    equipmentShortResponse `json:"equipment"`
	Materials    materialResponse       `json:"materials"`
}

// productTechCardResponse - технологическая карта продукта.
type productTechCardResponse struct {
	ProductID     int64                  `json:"product_id"`
	Product       *productShortResponse  `json:"product,omitempty"`
	TotalCost     float64                `json:"total_cost"`
	Operations    []operationResponse    `json:"operations"`
	MaterialNorms []materialNormResponse `json:"material_norms"`
}

// operationMaterialRequest - данные материала в запросе на создание операции.
type operationMaterialRequest struct {
	MaterialID     int64   `json:"material_id"`
	ProductID      int64   `json:"product_id"`
	MaterialAmount float32 `json:"material_amount"`
	Base           bool    `json:"base"`
}

// operationCreateRequest - тело запроса на создание операции.
type operationCreateRequest struct {
	Name         string                   `json:"name"`
	Description  string                   `json:"description,omitempty"`
	ClassID      int64                    `json:"operation_class_id"`
	ProfessionID int64                    `json:"profession_id"`
	Equipment    equipmentRequest         `json:"equipment"`
	StepNumber   int32                    `json:"step_number,omitempty"`
	LaborCost    float64                  `json:"labor_cost"`
	Parameters   map[string]any           `json:"parameters,omitempty"`
	Materials    operationMaterialRequest `json:"materials"`
}

func newOperationResponse(op entities.Operation) operationResponse {
	resp := operationResponse{
		ID:           op.ID,
		Name:         op.Name,
		Description:  op.Description,
		ClassID:      op.ClassID,
		ProfessionID: op.ProfessionID,
		StepNumber:   op.StepNumber,
		LaborCost:    op.LaborCost,
		Parameters:   op.Parameters,
		Equipment: equipmentShortResponse{
			Name:       op.Equipment.Name,
			WorkCenter: op.Equipment.WorkCenter,
		},
		Materials: materialResponse{
			MaterialID:     op.Materials.MaterialID,
			ProductID:      op.Materials.ProductID,
			MaterialAmount: op.Materials.MaterialAmount,
			Base:           op.Materials.Base,
		},
	}

	return resp
}

func newOperationsResponse(ops []entities.Operation) []operationResponse {
	resp := make([]operationResponse, 0, len(ops))
	for _, op := range ops {
		resp = append(resp, newOperationResponse(op))
	}
	return resp
}

func newProductTechCardResponse(tc entities.ProductTechCard) productTechCardResponse {
	ops := make([]operationResponse, 0, len(tc.Operations))
	for _, op := range tc.Operations {
		ops = append(ops, newOperationResponse(op))
	}

	norms := make([]materialNormResponse, 0, len(tc.MaterialNorms))
	for _, m := range tc.MaterialNorms {
		norms = append(norms, materialNormResponse{
			MaterialID:  m.MaterialID,
			Name:        m.Name,
			TotalAmount: m.MaterialAmount,
			Units:       m.Unit,
			Base:        m.Base,
		})
	}

	resp := productTechCardResponse{
		ProductID:     tc.ProductID,
		TotalCost:     tc.TotalCost,
		Operations:    ops,
		MaterialNorms: norms,
	}

	return resp
}

type operationHandlers struct {
	uc OperationService
}

// NewOperationHandlers создает новый экземпляр OperationHandlers.
func NewOperationHandlers(uc OperationService) OperationHandlers {
	return &operationHandlers{uc: uc}
}

// ListOperations обрабатывает GET /operations.
func (h *operationHandlers) ListOperations(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	var materialIDPtr *int64
	if raw := c.Query("material_id"); raw != "" {
		if id, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil && id > 0 {
			materialIDPtr = &id
		} else {
			badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("invalid material_id"))
			return
		}
	}

	var productIDPtr *int64
	if raw := c.Query("product_id"); raw != "" {
		if id, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil && id > 0 {
			productIDPtr = &id
		} else {
			badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("invalid product_id"))
			return
		}
	}

	ops, err := h.uc.GetOperations(c.Request.Context(), limit, offset, materialIDPtr, productIDPtr)
	if err != nil {
		handleError(c, log, "failed to list operations", err)
		return
	}

	c.JSON(http.StatusOK, newOperationsResponse(ops))
}

// CreateOperation обрабатывает POST /operations.
func (h *operationHandlers) CreateOperation(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req operationCreateRequest
	if !bindJSON(c, &req) {
		return
	}

	if req.Name == "" || req.ProfessionID <= 0 || req.Equipment.Name == "" || req.Equipment.WorkCenter == "" {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("name, profession_id and equipment (name, work_center) are required"))
		return
	}

	if req.Materials.MaterialID <= 0 || req.Materials.ProductID <= 0 {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("materials.material_id and materials.product_id are required"))
		return
	}

	opEntity := entities.Operation{
		Name:         req.Name,
		Description:  req.Description,
		ClassID:      req.ClassID,
		ProfessionID: req.ProfessionID,
		StepNumber:   req.StepNumber,
		LaborCost:    req.LaborCost,
		Parameters:   req.Parameters,
		Equipment: entities.Equipment{
			Name:       req.Equipment.Name,
			WorkCenter: req.Equipment.WorkCenter,
		},
		Materials: entities.MaterialProduct{
			Material: entities.Material{
				MaterialID:     req.Materials.MaterialID,
				ProductID:      req.Materials.ProductID,
				MaterialAmount: req.Materials.MaterialAmount,
				Base:           req.Materials.Base,
			},
		},
	}

	created, err := h.uc.CreateOperation(c.Request.Context(), opEntity)
	if err != nil {
		handleError(c, log, "failed to create operation", err)
		return
	}

	c.JSON(http.StatusCreated, newOperationResponse(*created))
}

// GetOperation обрабатывает GET /operations/{id}.
func (h *operationHandlers) GetOperation(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	op, err := h.uc.GetOperationByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, log, "failed to get operation", err)
		return
	}

	c.JSON(http.StatusOK, newOperationResponse(*op))
}

// UpdateOperation обрабатывает PATCH /operations/{id}.
func (h *operationHandlers) UpdateOperation(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	var updates map[string]any
	decoder := json.NewDecoder(c.Request.Body)
	decoder.UseNumber()
	if decodeErr := decoder.Decode(&updates); decodeErr != nil {
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidJSON, "Request body is not valid JSON")
		return
	}
	if len(updates) == 0 {
		writeError(c, http.StatusBadRequest, apperrors.CodeBadRequest, "Empty update payload")
		return
	}

	updated, err := h.uc.UpdateOperation(c.Request.Context(), id, updates)
	if err != nil {
		handleError(c, log, "failed to update operation", err)
		return
	}

	c.JSON(http.StatusOK, newOperationResponse(*updated))
}

// DeleteOperation обрабатывает DELETE /operations/{id}.
func (h *operationHandlers) DeleteOperation(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	if err := h.uc.DeleteOperation(c.Request.Context(), id); err != nil {
		handleError(c, log, "failed to delete operation", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProductTechCard обрабатывает GET /products/{id}/operations.
func (h *operationHandlers) GetProductTechCard(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	productID, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	quantity, err := getPositiveFloatQuery(c, "quantity", 1)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	techCard, err := h.uc.GetProductTechCard(c.Request.Context(), productID, quantity)
	if err != nil {
		handleError(c, log, "failed to get product tech card", err)
		return
	}

	c.JSON(http.StatusOK, newProductTechCardResponse(*techCard))
}
