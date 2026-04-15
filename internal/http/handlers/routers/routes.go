// Package routers содержит регистрацию все url путей сервера
package routers

import (
	"github.com/gin-gonic/gin"
	"Brewery/internal/http/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(e *gin.Engine, categoryHandler handlers.CategoriesHandlers, beersHandler handlers.BeersHandlers) {
	api := e.Group("/api")
	{
		api.POST("/beers", beersHandler.CreateBeer)
		api.PATCH("/beers/:id", beersHandler.UpdateBeer)
		api.DELETE("/beers/:id", beersHandler.DeleteBeer)
		api.GET("/beers", beersHandler.GetAllBeers)
		api.POST("/beers/reviews/:beer_id", beersHandler.CreateBeerReview)

		api.POST("/categories", categoryHandler.CreateCategory)
		api.GET("/categories/:id", categoryHandler.GetCategoryById)
		api.PATCH("/categories/:id", categoryHandler.UpdateCategory)
		api.DELETE("/categories/:id", categoryHandler.DeleteCategory)
		api.GET("/categories", categoryHandler.GetAllCategories)
		api.GET("/categories/beers/:category_id", categoryHandler.GetBeersByCategory)
		api.GET("/categories/parent/:id", categoryHandler.GetParentCategory)
		api.GET("/categories/children/:id", categoryHandler.GetChildCategory)
	}
		
	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
}


