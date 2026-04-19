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
		beers := api.Group("/beers")
		{
			beers.POST("", beersHandler.CreateBeer)
			beers.PATCH("/:id", beersHandler.UpdateBeer)
			beers.DELETE("/:id", beersHandler.DeleteBeer)
			beers.GET("", beersHandler.GetAllBeers)
			beers.POST("/reviews/:beer_id", beersHandler.CreateBeerReview)
		}

		reviews := api.Group("/reviews")
		{
			reviews.POST("/:beer_id", beersHandler.CreateBeerReview)
		}
		
		categories := api.Group("/categories")
		{
			categories.POST("", categoryHandler.CreateCategory)
			categories.GET("/:id", categoryHandler.GetCategoryByID)
			categories.PATCH("/:id", categoryHandler.UpdateCategory)
			categories.DELETE("/:id", categoryHandler.DeleteCategory)
			categories.GET("", categoryHandler.GetAllCategories)
			categories.GET("/beers/:category_id", categoryHandler.GetBeersByCategory)
			categories.GET("/parent/:id", categoryHandler.GetParentCategory)
			categories.GET("/children/:id", categoryHandler.GetChildCategory)
		}
	}
		
	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
}


