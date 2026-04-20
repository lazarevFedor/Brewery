// Package routers содержит регистрацию все url путей сервера
package routers

import (
	"Brewery/internal/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(
	e *gin.Engine,
	categoryHandler handlers.CategoriesHandlers,
	beersHandler handlers.BeersHandlers,
	reviewHandler handlers.ReviewsHandlers,
) {
	api := e.Group("/api")
	{
		beers := api.Group("/beers")
		{
			beers.POST("", beersHandler.CreateBeer)
			beers.PATCH("/:id", beersHandler.UpdateBeer)
			beers.DELETE("/:id", beersHandler.DeleteBeer)
			beers.GET("", beersHandler.GetAllBeers)

			features := beers.Group("/feats")
			{
				features.GET("/:beer_id", beersHandler.GetFeature)
				features.POST("/:beer_id", beersHandler.CreateFeature)
				features.PATCH("/:beer_id", beersHandler.UpdateBeer)
				features.DELETE("/:beer_id", beersHandler.DeleteFeature)
			}
		}

		reviews := api.Group("/reviews")
		{
			reviews.POST("/:beer_id", reviewHandler.CreateReview)
			reviews.GET("/:beer_id", reviewHandler.GetBeersReviews)
			reviews.DELETE("/:id", reviewHandler.DeleteReview)
			reviews.PATCH("/:id", reviewHandler.UpdateReview)
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
