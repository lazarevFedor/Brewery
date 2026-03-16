package handlers

import "github.com/gin-gonic/gin"

type beerClassifier struct {
}

type BeersHandler struct {
	beerCl beerClassifier
}

func NewBeersHandler(bc beerClassifier) *BeersHandler {
	return &BeersHandler{
		beerCl: bc,
	}
}

func (h BeersHandler) Create(_ *gin.Context) {}
func (h BeersHandler) Get(_ *gin.Context)    {}
func (h BeersHandler) Update(_ *gin.Context) {}
func (h BeersHandler) Delete(_ *gin.Context) {}
func (h BeersHandler) List(_ *gin.Context)   {}
