package handlers

import (
	"Brewery/internal/usecase"

	"github.com/gin-gonic/gin"
)

type EnumClassHandlers interface {
	CreateEnum(c *gin.Context)
	GetEnum(c *gin.Context)
	UpdateEnum(c *gin.Context)
	DeleteEnum(c *gin.Context)
}

type enumClassHandlers struct {
	uc usecase.BeerService
}

func NewEnumClassHandlers(usecase usecase.BeerService) EnumClassHandlers {
	return &enumClassHandlers{
		uc: usecase,
	}
}

func (h *enumClassHandlers) CreateEnum(c *gin.Context) {

}

func (h *enumClassHandlers) GetEnum(c *gin.Context) {

}

func (h *enumClassHandlers) UpdateEnum(c *gin.Context) {

}

func (h *enumClassHandlers) DeleteEnum(c *gin.Context) {

}
