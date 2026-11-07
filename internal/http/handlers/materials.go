package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MaterialService определяет бизнес-логику работы с составом изделий.
// Сигнатуры совпадают с usecase.MaterialService, кроме GetMaterialNorms - его там еще нет.
type MaterialService interface {

	// CreateMaterial добавляет материал в состав изделия.
	CreateMaterial(ctx context.Context, material entities.Material) (*entities.Material, error)

	// UpdateMaterial меняет количество материала в составе изделия.
	UpdateMaterial(ctx context.Context, productID, materialID int64, amount float32) (*entities.Material, error)

	// DeleteMaterial убирает материал из состава изделия.
	DeleteMaterial(ctx context.Context, productID, materialID int64) error

	// GetMaterialsByProduct возвращает полное дерево состава изделия на всю глубину (FIND_BOM).
	// Вложенный состав каждого материала лежит в MaterialProduct.Materials.
	GetMaterialsByProduct(ctx context.Context, productID int64) ([]entities.MaterialProduct, error)

	// GetMaterialNorms возвращает сводные нормы расхода материалов заданного класса для изделия:
	// каждый материал один раз, в MaterialAmount - сколько всего его нужно с учетом всей глубины состава.
	GetMaterialNorms(ctx context.Context, productID, classID int64) ([]entities.MaterialProduct, error)

	// GetProductsByMaterial возвращает изделия, в состав которых входит материал.
	GetProductsByMaterial(ctx context.Context, materialID int64, limit, offset int32) ([]entities.MaterialProduct, error)
}

// MaterialHandlers - HTTP-обработчики для состава изделий.
type MaterialHandlers interface {

	// GetProductBomTree обрабатывает GET /products/{id}/materials.
	GetProductBomTree(c *gin.Context)

	// AddProductMaterial обрабатывает POST /products/{id}/materials.
	AddProductMaterial(c *gin.Context)

	// UpdateProductMaterial обрабатывает PATCH /products/{id}/materials/{material_id}.
	UpdateProductMaterial(c *gin.Context)

	// DeleteProductMaterial обрабатывает DELETE /products/{id}/materials/{material_id}.
	DeleteProductMaterial(c *gin.Context)

	// CalculateMaterialNorms обрабатывает GET /products/{id}/materials/norms.
	CalculateMaterialNorms(c *gin.Context)

	// ListProductsByMaterial обрабатывает GET /materials/{material_id}/products.
	ListProductsByMaterial(c *gin.Context)
}

// materialResponse - представление материала в составе изделия (схема Material в swagger).
type materialResponse struct {
	MaterialID     int64   `json:"material_id"`
	ProductID      int64   `json:"product_id"`
	MaterialAmount float32 `json:"material_amount"`
}

func newMaterialResponse(m entities.Material) materialResponse {
	return materialResponse{
		MaterialID:     m.MaterialID,
		ProductID:      m.ProductID,
		MaterialAmount: m.MaterialAmount,
	}
}

// materialTreeNodeResponse - узел дерева состава (схема MaterialTreeNode в swagger).
type materialTreeNodeResponse struct {
	MaterialID     int64                      `json:"material_id"`
	Name           string                     `json:"name"`
	MaterialAmount float32                    `json:"material_amount"`
	Units          string                     `json:"units,omitempty"`
	Materials      []materialTreeNodeResponse `json:"materials"`
}

func newMaterialTreeResponse(items []entities.MaterialProduct) []materialTreeNodeResponse {
	resp := make([]materialTreeNodeResponse, 0, len(items))
	for _, m := range items {
		resp = append(resp, materialTreeNodeResponse{
			MaterialID:     m.ID,
			Name:           m.Name,
			MaterialAmount: m.MaterialAmount,
			Units:          m.Unit,
			Materials:      newMaterialTreeResponse(m.Materials),
		})
	}

	return resp
}

// materialNormResponse - норма расхода материала (схема MaterialNorm в swagger).
type materialNormResponse struct {
	MaterialID  int64   `json:"material_id"`
	Name        string  `json:"name"`
	TotalAmount float32 `json:"total_amount"`
	Units       string  `json:"units,omitempty"`
}

// materialCreateRequest - тело POST /products/{id}/materials (схема MaterialCreate).
type materialCreateRequest struct {
	MaterialID     int64   `json:"material_id"`
	MaterialAmount float32 `json:"material_amount"`
}

// materialUpdateRequest - тело PATCH /products/{id}/materials/{material_id} (схема MaterialUpdate).
type materialUpdateRequest struct {
	MaterialAmount float32 `json:"material_amount"`
}

// materialHandlers - реализация MaterialHandlers.
type materialHandlers struct {
	uc MaterialService
}

// NewMaterialHandlers создает обработчики состава изделий. Если uc == nil, все запросы отвечают 501.
func NewMaterialHandlers(uc MaterialService) MaterialHandlers {
	return &materialHandlers{uc: uc}
}

// GetProductBomTree обрабатывает GET /products/{id}/materials.
func (h *materialHandlers) GetProductBomTree(c *gin.Context) {
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

	tree, err := h.uc.GetMaterialsByProduct(c.Request.Context(), productID)
	if err != nil {
		handleError(c, log, "failed to get product composition tree", err)
		return
	}

	c.JSON(http.StatusOK, newMaterialTreeResponse(tree))
}

// AddProductMaterial обрабатывает POST /products/{id}/materials.
func (h *materialHandlers) AddProductMaterial(c *gin.Context) {
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

	var req materialCreateRequest
	if !bindJSON(c, &req) {
		return
	}

	if req.MaterialID <= 0 {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "material_id is required")
		return
	}

	if req.MaterialID == productID {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "Product cannot be a material of itself")
		return
	}

	if req.MaterialAmount <= 0 {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "material_amount must be greater than 0")
		return
	}

	created, err := h.uc.CreateMaterial(c.Request.Context(), entities.Material{
		MaterialID:     req.MaterialID,
		ProductID:      productID,
		MaterialAmount: req.MaterialAmount,
	})
	if err != nil {
		handleError(c, log, "failed to add material", err)
		return
	}

	log.Info(c.Request.Context(), "action=create resource=material status=success",
		zap.Int64("product_id", productID), zap.Int64("material_id", req.MaterialID))
	c.JSON(http.StatusCreated, newMaterialResponse(*created))
}

// UpdateProductMaterial обрабатывает PATCH /products/{id}/materials/{material_id}.
func (h *materialHandlers) UpdateProductMaterial(c *gin.Context) {
	if h.uc == nil {
		notImplemented(c)
		return
	}

	log, ok := getLogger(c)
	if !ok {
		return
	}

	productID, materialID, ok := h.getMaterialPath(c)
	if !ok {
		return
	}

	var req materialUpdateRequest
	if !bindJSON(c, &req) {
		return
	}

	if req.MaterialAmount <= 0 {
		writeError(c, http.StatusBadRequest, apperrors.CodeValidation, "material_amount must be greater than 0")
		return
	}

	updated, err := h.uc.UpdateMaterial(c.Request.Context(), productID, materialID, req.MaterialAmount)
	if err != nil {
		handleError(c, log, "failed to update material", err)
		return
	}

	log.Info(c.Request.Context(), "action=update resource=material status=success",
		zap.Int64("product_id", productID), zap.Int64("material_id", materialID))
	c.JSON(http.StatusOK, newMaterialResponse(*updated))
}

// DeleteProductMaterial обрабатывает DELETE /products/{id}/materials/{material_id}.
func (h *materialHandlers) DeleteProductMaterial(c *gin.Context) {
	if h.uc == nil {
		notImplemented(c)
		return
	}

	log, ok := getLogger(c)
	if !ok {
		return
	}

	productID, materialID, ok := h.getMaterialPath(c)
	if !ok {
		return
	}

	if err := h.uc.DeleteMaterial(c.Request.Context(), productID, materialID); err != nil {
		handleError(c, log, "failed to delete material", err)
		return
	}

	log.Info(c.Request.Context(), "action=delete resource=material status=success",
		zap.Int64("product_id", productID), zap.Int64("material_id", materialID))
	c.Status(http.StatusNoContent)
}

// CalculateMaterialNorms обрабатывает GET /products/{id}/materials/norms?class_id=.
func (h *materialHandlers) CalculateMaterialNorms(c *gin.Context) {
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

	classID, err := getIDQuery(c, "class_id", true)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	norms, err := h.uc.GetMaterialNorms(c.Request.Context(), productID, classID)
	if err != nil {
		handleError(c, log, "failed to calculate material norms", err)
		return
	}

	resp := make([]materialNormResponse, 0, len(norms))
	for _, n := range norms {
		resp = append(resp, materialNormResponse{
			MaterialID:  n.ID,
			Name:        n.Name,
			TotalAmount: n.MaterialAmount,
			Units:       n.Unit,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// ListProductsByMaterial обрабатывает GET /materials/{material_id}/products.
func (h *materialHandlers) ListProductsByMaterial(c *gin.Context) {
	if h.uc == nil {
		notImplemented(c)
		return
	}

	log, ok := getLogger(c)
	if !ok {
		return
	}

	materialID, err := getIDParam(c, "material_id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return
	}

	offset, limit, err := getPaginationParams(c)
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidParameters, err)
		return
	}

	usages, err := h.uc.GetProductsByMaterial(c.Request.Context(), materialID, limit, offset)
	if err != nil {
		handleError(c, log, "failed to list products by material", err)
		return
	}

	products := make([]productResponse, 0, len(usages))
	for _, u := range usages {
		products = append(products, newProductResponse(u.Product))
	}

	c.JSON(http.StatusOK, products)
}

// getMaterialPath разбирает id изделия и material_id из пути.
func (h *materialHandlers) getMaterialPath(c *gin.Context) (productID, materialID int64, ok bool) {
	log, ok := getLogger(c)
	if !ok {
		return 0, 0, false
	}

	productID, err := getIDParam(c, "id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return 0, 0, false
	}

	materialID, err = getIDParam(c, "material_id")
	if err != nil {
		badRequest(c, log, apperrors.CodeInvalidID, err)
		return 0, 0, false
	}

	return productID, materialID, true
}
