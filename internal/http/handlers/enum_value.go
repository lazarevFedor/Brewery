package handlers

import (
	"Brewery/internal/usecase"

	"github.com/gin-gonic/gin"
)

type EnumValueHandlers interface {
	CreateValue(c *gin.Context)
	GetValue(c *gin.Context)
	UpdateValue(c *gin.Context)
	DeleteValue(c *gin.Context)
}

type enumValueHandlers struct {
	uc usecase.BeerService
}

func NewEnumValueHandlers(usecase usecase.BeerService) EnumValueHandlers {
	return &enumValueHandlers{
		uc: usecase,
	}
}

func (h *enumValueHandlers) CreateValue(c *gin.Context) {

}

func (h *enumValueHandlers) GetValue(c *gin.Context) {

}

func (h *enumValueHandlers) UpdateValue(c *gin.Context) {

}

func (h *enumValueHandlers) DeleteValue(c *gin.Context) {

}
