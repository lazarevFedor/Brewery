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
	// CreateOperation создает операцию. Если передан новый входной/выходной продукт (inputProduct/outputProduct != nil),
	// сервис автоматически создает его в базе данных.
	CreateOperation(ctx context.Context, op entities.Operation, inputProduct, outputProduct *entities.Product) (*entities.Operation, error)

	// GetOperations возвращает список операций с учетом пагинации и опциональной фильтрации по продуктам.
	GetOperations(ctx context.Context, limit, offset int32, inputProductID, outputProductID *int64) ([]entities.Operation, error)

	// GetOperationByID возвращает операцию по ID с заполненными связями (оборудование, продукты).
	GetOperationByID(ctx context.Context, id int64) (*entities.Operation, error)

	// UpdateOperation обновляет операцию.
	UpdateOperation(ctx context.Context, id int64, updates map[string]any, inputProduct, outputProduct *entities.Product) (*entities.Operation, error)

	// DeleteOperation удаляет операцию по идентификатору.
	DeleteOperation(ctx context.Context, id int64) error

	// GetProductTechCard возвращает технологическую карту продукта (операции, суммарные человеко-часы, нормы расхода).
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
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// operationResponse - представление операции в API.
type operationResponse struct {
	ID            int64                  `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	ProfessionID  int64                  `json:"profession_id"`
	Equipment     equipmentShortResponse `json:"equipment"`
	InputProduct  productShortResponse   `json:"input_product"`
	OutputProduct productShortResponse   `json:"output_product"`
	LaborCost     float64                `json:"labor_cost"`
	Parameters    map[string]any         `json:"parameters,omitempty"`
}

// productTechCardResponse - технологическая карта продукта.
type productTechCardResponse struct {
	ProductID      int64                  `json:"product_id"`
	Product        *productShortResponse  `json:"product,omitempty"`
	TotalLaborCost float64                `json:"total_labor_cost"`
	Operations     []operationResponse    `json:"operations"`
	MaterialNorms  []materialNormResponse `json:"material_norms"`
}

// operationCreateRequest - тело запроса на создание операции.
type operationCreateRequest struct {
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	ProfessionID    int64           `json:"profession_id"`
	EquipmentID     int64           `json:"equipment_id"`
	InputProductID  *int64          `json:"input_product_id"`
	InputProduct    json.RawMessage `json:"input_product"`
	OutputProductID *int64          `json:"output_product_id"`
	OutputProduct   json.RawMessage `json:"output_product"`
	LaborCost       float64         `json:"labor_cost"`
	Parameters      map[string]any  `json:"parameters"`
}

func parseProductOrID(raw json.RawMessage, directID *int64) (int64, *entities.Product, error) {
	if directID != nil && *directID > 0 {
		return *directID, nil, nil
	}

	if len(raw) == 0 {
		return 0, nil, errors.New("product is required")
	}

	// Если передан числовой ID в поле product
	var numID int64
	if err := json.Unmarshal(raw, &numID); err == nil && numID > 0 {
		return numID, nil, nil
	}

	// Иначе пробуем распарсить схему ProductCreate для создания нового продукта
	var pReq productCreateRequest
	if err := json.Unmarshal(raw, &pReq); err != nil {
		return 0, nil, errors.New("invalid product data")
	}

	if pReq.Name == "" {
		return 0, nil, errors.New("product name is required")
	}

	entity := pReq.toEntity()
	return 0, &entity, nil
}

func newOperationResponse(op entities.Operation) operationResponse {
	resp := operationResponse{
		ID:           op.ID,
		Name:         op.Name,
		Description:  op.Description,
		ProfessionID: op.ProfessionID,
		LaborCost:    op.LaborCost,
		Parameters:   op.Parameters,
		Equipment: equipmentShortResponse{
			ID: op.EquipmentID,
		},
		InputProduct: productShortResponse{
			ID: op.InputProductID,
		},
		OutputProduct: productShortResponse{
			ID: op.OutputProductID,
		},
	}

	if op.Equipment != nil {
		resp.Equipment.Name = op.Equipment.Name
	}
	if op.InputProduct != nil {
		resp.InputProduct.Name = op.InputProduct.Name
		resp.InputProduct.IsFinalProduct = op.InputProduct.IsFinalProduct
	}
	if op.OutputProduct != nil {
		resp.OutputProduct.Name = op.OutputProduct.Name
		resp.OutputProduct.IsFinalProduct = op.OutputProduct.IsFinalProduct
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
		ProductID:      tc.ProductID,
		TotalLaborCost: tc.TotalLaborCost,
		Operations:     ops,
		MaterialNorms:  norms,
	}

	if tc.Product != nil {
		resp.Product = &productShortResponse{
			ID:             tc.Product.ID,
			Name:           tc.Product.Name,
			IsFinalProduct: tc.Product.IsFinalProduct,
		}
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

	var inputIDPtr *int64
	if raw := c.Query("input_product_id"); raw != "" {
		if id, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil && id > 0 {
			inputIDPtr = &id
		} else {
			badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("invalid input_product_id"))
			return
		}
	}

	var outputIDPtr *int64
	if raw := c.Query("output_product_id"); raw != "" {
		if id, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil && id > 0 {
			outputIDPtr = &id
		} else {
			badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("invalid output_product_id"))
			return
		}
	}

	ops, err := h.uc.GetOperations(c.Request.Context(), limit, offset, inputIDPtr, outputIDPtr)
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

	if req.Name == "" || req.Description == "" || req.ProfessionID <= 0 || req.EquipmentID <= 0 {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("name, description, profession_id and equipment_id are required"))
		return
	}

	inputID, inputToCreate, err := parseProductOrID(req.InputProduct, req.InputProductID)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("invalid input product: "+err.Error()))
		return
	}

	outputID, outputToCreate, err := parseProductOrID(req.OutputProduct, req.OutputProductID)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("invalid output product: "+err.Error()))
		return
	}

	opEntity := entities.Operation{
		Name:            req.Name,
		Description:     req.Description,
		ProfessionID:    req.ProfessionID,
		EquipmentID:     req.EquipmentID,
		InputProductID:  inputID,
		OutputProductID: outputID,
		LaborCost:       req.LaborCost,
		Parameters:      req.Parameters,
	}

	created, err := h.uc.CreateOperation(c.Request.Context(), opEntity, inputToCreate, outputToCreate)
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

	updates, ok := bindUpdates(c)
	if !ok {
		return
	}

	var inputToCreate *entities.Product
	if raw, exists := updates["input_product"]; exists {
		if rawMap, isMap := raw.(map[string]any); isMap {
			rawJSON, _ := json.Marshal(rawMap)
			_, p, parseErr := parseProductOrID(rawJSON, nil)
			if parseErr != nil {
				badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("invalid input_product: "+parseErr.Error()))
				return
			}
			inputToCreate = p
			delete(updates, "input_product")
		} else if num, isNum := raw.(float64); isNum {
			updates["input_product_id"] = int64(num)
			delete(updates, "input_product")
		}
	}

	var outputToCreate *entities.Product
	if raw, exists := updates["output_product"]; exists {
		if rawMap, isMap := raw.(map[string]any); isMap {
			rawJSON, _ := json.Marshal(rawMap)
			_, p, parseErr := parseProductOrID(rawJSON, nil)
			if parseErr != nil {
				badRequest(c, log, apperrors.CodeInvalidParameters, errors.New("invalid output_product: "+parseErr.Error()))
				return
			}
			outputToCreate = p
			delete(updates, "output_product")
		} else if num, isNum := raw.(float64); isNum {
			updates["output_product_id"] = int64(num)
			delete(updates, "output_product")
		}
	}

	updated, err := h.uc.UpdateOperation(c.Request.Context(), id, updates, inputToCreate, outputToCreate)
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
