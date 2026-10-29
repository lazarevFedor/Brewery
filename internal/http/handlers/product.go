package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/pkg/logger"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type productService interface {
	CreateProduct(ctx context.Context, req entities.Product) (entities.Product, error)
	GetProducts(ctx context.Context, limit, offset uint64) ([]entities.Product, error)
	UpdateProduct(ctx context.Context, id uint, updates map[string]any) (entities.Product, error)
	DeleteProduct(ctx context.Context, id int) error
	GetProductsByClass(ctx context.Context, classID int, limit, offset uint64) ([]entities.Product, error)
	
}

type ProductHandler struct{
	service productService
}

func NewProductHandler(service productService) *ProductHandler {
	return &ProductHandler{service: service}
}


// CreateProduct обрабатывает HTTP-запрос на добавление нового продукта.
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	var req entities.Product
	
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(c.Request.Context(), "failed to Unmurshal JSON", zap.Error(err))
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidJSON, "Request body is not valid JSON")
		return
	}

	prod, err := h.service.CreateProduct(c.Request.Context(), req)
	if err != nil {
		log.Error(c.Request.Context(), fmt.Sprintf("Failed to create product: %v", err))
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")

		return
	}

	log.Debug(c.Request.Context(), fmt.Sprintf("product_id = %d", prod.ID))
	log.Info(c.Request.Context(), fmt.Sprintf("action=create resource=beer status=success name=%q", req.Name))
	c.JSON(http.StatusCreated, prod)
}
