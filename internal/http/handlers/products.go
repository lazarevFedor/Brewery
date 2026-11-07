package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProductService определяет бизнес-логику работы с изделиями, которая нужна HTTP-слою.
type ProductService interface {
	CreateProduct(ctx context.Context, product entities.Product) (*entities.Product, error)
	GetProducts(ctx context.Context, limit, offset int32) ([]entities.Product, error)
	GetProductByID(ctx context.Context, id int64) (*entities.Product, error)
	UpdateProduct(ctx context.Context, id int64, updates map[string]any) (*entities.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProductsByClass(ctx context.Context, classID int64, limit, offset int32) ([]entities.Product, error)
	GetFilteredProducts(ctx context.Context, filters []entities.FilterParameter, classID *uint, limit, offset int32) ([]entities.Product, error)
}

// DerivedProductService возвращает изделия, созданные на основе базового (base_id = id).
type DerivedProductService interface {
	GetDerivedProducts(ctx context.Context, baseID int64, limit, offset int32) ([]entities.Product, error)
}

// ProductHandlers определяет HTTP-обработчики для изделий.
type ProductHandlers interface {

	// ListProducts обрабатывает GET /products.
	ListProducts(c *gin.Context)

	// CreateProduct обрабатывает POST /products.
	CreateProduct(c *gin.Context)

	// GetProduct обрабатывает GET /products/{id}.
	GetProduct(c *gin.Context)

	// UpdateProduct обрабатывает PATCH /products/{id}.
	UpdateProduct(c *gin.Context)

	// DeleteProduct обрабатывает DELETE /products/{id}.
	DeleteProduct(c *gin.Context)

	// ListProductsByClass обрабатывает GET /classes/products/{class_id}.
	ListProductsByClass(c *gin.Context)

	// SearchProducts обрабатывает GET /products/search.
	SearchProducts(c *gin.Context)

	// SearchProductsByClass обрабатывает GET /classes/{id}/products/search.
	SearchProductsByClass(c *gin.Context)

	// ListDerivedProducts обрабатывает GET /products/{id}/derived.
	ListDerivedProducts(c *gin.Context)
}

// productResponse - представление изделия в API (схема Product в swagger).
type productResponse struct {
	ID         int64          `json:"id"`
	Name       string         `json:"name"`
	Class      classBase      `json:"class"`
	ClassID    int64          `json:"class_id"`
	BaseID     *int64         `json:"base_id"`
	Parameters map[string]any `json:"parameters,omitempty"`
	Amount     float32        `json:"amount,omitempty"`
	Units      string         `json:"units,omitempty"`
}

// classBase - краткое представление класса изделия (схема ClassBase в swagger).
type classBase struct {
	Name string `json:"name"`
}

// productCreateRequest - тело запроса на создание изделия (схема ProductCreate в swagger).
type productCreateRequest struct {
	Name       string         `json:"name"`
	Class      *classBase     `json:"class"`
	ClassID    int64          `json:"class_id"`
	BaseID     *int64         `json:"base_id"`
	Parameters map[string]any `json:"parameters"`
	Amount     float32        `json:"amount"`
	Units      string         `json:"units"`
}

func (r productCreateRequest) toEntity() entities.Product {
	product := entities.Product{
		Name:       r.Name,
		ProdClass:  entities.Class{ID: r.ClassID},
		Parameters: r.Parameters,
		Amount:     r.Amount,
		Unit:       r.Units,
	}

	if r.Class != nil {
		product.ProdClass.Name = r.Class.Name
	}

	if r.BaseID != nil {
		product.BaseID = *r.BaseID
	}

	return product
}

func newProductResponse(p entities.Product) productResponse {
	resp := productResponse{
		ID:         p.ID,
		Name:       p.Name,
		Class:      classBase{Name: p.ProdClass.Name},
		ClassID:    p.ProdClass.ID,
		Parameters: p.Parameters,
		Amount:     p.Amount,
		Units:      p.Unit,
	}

	if p.BaseID != 0 {
		baseID := p.BaseID
		resp.BaseID = &baseID
	}

	return resp
}

func newProductsResponse(products []entities.Product) []productResponse {
	resp := make([]productResponse, 0, len(products))
	for _, p := range products {
		resp = append(resp, newProductResponse(p))
	}

	return resp
}

// productUpdateIntFields - поля PATCH-запроса, которые приходят из JSON как float64, а сервис ждет целое.
var productUpdateIntFields = []string{"class_id", "base_id"}

// productHandlers реализует ProductHandlers.
type productHandlers struct {
	uc      ProductService
	derived DerivedProductService
}

// NewProductHandlers создает обработчики изделий. derived может быть nil - тогда /products/{id}/derived отвечает 501.
func NewProductHandlers(uc ProductService, derived DerivedProductService) ProductHandlers {
	return &productHandlers{uc: uc, derived: derived}
}

// ListProducts обрабатывает GET /products.
func (h *productHandlers) ListProducts(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	products, err := h.uc.GetProducts(c.Request.Context(), limit, offset)
	if err != nil {
		handleError(c, log, "failed to list products", err)
		return
	}

	c.JSON(http.StatusOK, newProductsResponse(products))
}

// CreateProduct обрабатывает POST /products.
func (h *productHandlers) CreateProduct(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	var req productCreateRequest
	if !bindJSON(c, &req) {
		return
	}

	if req.Name == "" {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "Product name is required")
		return
	}

	if req.ClassID == 0 && (req.Class == nil || req.Class.Name == "") {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "Product class is required")
		return
	}

	created, err := h.uc.CreateProduct(c.Request.Context(), req.toEntity())
	if err != nil {
		handleError(c, log, "failed to create product", err)
		return
	}

	log.Info(c.Request.Context(), "action=create resource=product status=success", zap.Int64("id", created.ID))
	c.JSON(http.StatusCreated, newProductResponse(*created))
}

// GetProduct обрабатывает GET /products/{id}.
func (h *productHandlers) GetProduct(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	product, err := h.uc.GetProductByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, log, "failed to get product", err)
		return
	}

	if product == nil {
		writeError(c, http.StatusNotFound, apperrors.CodeNotFound, "Product not found")
		return
	}

	c.JSON(http.StatusOK, newProductResponse(*product))
}

// UpdateProduct обрабатывает PATCH /products/{id}.
func (h *productHandlers) UpdateProduct(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	updates, ok := bindUpdates(c)
	if !ok {
		return
	}

	for _, field := range productUpdateIntFields {
		raw, exists := updates[field]
		if !exists {
			continue
		}

		num, isNum := raw.(float64)
		if !isNum || num != float64(int64(num)) {
			writeError(c, http.StatusBadRequest, apperrors.CodeValidation, field+" must be an integer")
			return
		}

		updates[field] = int64(num)
	}

	updated, err := h.uc.UpdateProduct(c.Request.Context(), id, updates)
	if err != nil {
		handleError(c, log, "failed to update product", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=product status=success", zap.Int64("id", id))
	c.JSON(http.StatusOK, newProductResponse(*updated))
}

// DeleteProduct обрабатывает DELETE /products/{id}.
func (h *productHandlers) DeleteProduct(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	id, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	if err = h.uc.DeleteProduct(c.Request.Context(), id); err != nil {
		handleError(c, log, "failed to delete product", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=product status=success", zap.Int64("id", id))
	c.Status(http.StatusNoContent)
}

// ListProductsByClass обрабатывает GET /classes/products/{class_id}.
func (h *productHandlers) ListProductsByClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classID, err := getIDParam(c, "class_id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	products, err := h.uc.GetProductsByClass(c.Request.Context(), classID, limit, offset)
	if err != nil {
		handleError(c, log, "failed to list products by class", err)
		return
	}

	c.JSON(http.StatusOK, newProductsResponse(products))
}

// SearchProducts обрабатывает GET /products/search. По swagger здесь только точное совпадение (eq).
func (h *productHandlers) SearchProducts(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	filters, err := getFilters(c, []entities.Operation{entities.OpEqual})
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	h.search(c, filters, nil)
}

// SearchProductsByClass обрабатывает GET /classes/{id}/products/search. Здесь разрешены и диапазоны.
func (h *productHandlers) SearchProductsByClass(c *gin.Context) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	classID, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	filters, err := getFilters(c, nil)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	id := toUint(classID)
	h.search(c, filters, &id)
}

func (h *productHandlers) search(c *gin.Context, filters []entities.FilterParameter, classID *uint) {
	log, ok := getLogger(c)
	if !ok {
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	products, err := h.uc.GetFilteredProducts(c.Request.Context(), filters, classID, limit, offset)
	if err != nil {
		handleError(c, log, "failed to search products", err)
		return
	}

	c.JSON(http.StatusOK, newProductsResponse(products))
}

// ListDerivedProducts обрабатывает GET /products/{id}/derived.
func (h *productHandlers) ListDerivedProducts(c *gin.Context) {
	if h.derived == nil {
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

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	products, err := h.derived.GetDerivedProducts(c.Request.Context(), id, limit, offset)
	if err != nil {
		handleError(c, log, "failed to list derived products", err)
		return
	}

	c.JSON(http.StatusOK, newProductsResponse(products))
}
