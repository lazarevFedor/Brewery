// Package routers содержит регистрацию все url путей сервера
package routers

import (
	"Brewery/internal/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(e *gin.Engine, h handlers.Handlers) {
	api := e.Group("/api")
	{
		beers := api.Group("/beers")
		{
			beers.POST("", h.BeersHandler.CreateBeer)
			beers.PATCH("/:id", h.BeersHandler.UpdateBeer)
			beers.DELETE("/:id", h.BeersHandler.DeleteBeer)
			beers.GET("", h.BeersHandler.GetAllBeers)

			features := beers.Group("/feats")
			{
				features.GET("/:beer_id", h.BeersHandler.GetFeature)
				features.POST("/:beer_id", h.BeersHandler.CreateFeature)
				features.PATCH("/:beer_id", h.BeersHandler.UpdateBeer)
				features.DELETE("/:beer_id", h.BeersHandler.DeleteFeature)
			}
		}

		reviews := api.Group("/reviews")
		{
			reviews.POST("/:beer_id", h.ReviewHandler.CreateReview)
			reviews.GET("/:beer_id", h.ReviewHandler.GetBeersReviews)
			reviews.DELETE("/:id", h.ReviewHandler.DeleteReview)
			reviews.PATCH("/:id", h.ReviewHandler.UpdateReview)
		}

		categories := api.Group("/categories")
		{
			categories.POST("", h.CategoryHandler.CreateCategory)
			categories.GET("/:id", h.CategoryHandler.GetCategoryByID)
			categories.PATCH("/:id", h.CategoryHandler.UpdateCategory)
			categories.DELETE("/:id", h.CategoryHandler.DeleteCategory)
			categories.GET("", h.CategoryHandler.GetAllCategories)
			categories.GET("/beers/:category_id", h.CategoryHandler.GetBeersByCategory)
			categories.GET("/parent/:id", h.CategoryHandler.GetParentCategory)
			categories.GET("/children/:id", h.CategoryHandler.GetChildCategory)
		}

		enums := api.Group("/enums")
		{
			enums.POST("", h.EnumClassHandler.CreateEnum)
			enums.GET("", h.EnumClassHandler.GetEnum)
			enums.PATCH("/:id", h.EnumClassHandler.UpdateEnum)
			enums.DELETE("/:id", h.EnumClassHandler.DeleteEnum)

			value := enums.Group("value")
			{
				value.POST("", h.EnumValueHandler.CreateValue)
				value.GET("", h.EnumValueHandler.GetValue)
				value.PATCH("/:id", h.EnumValueHandler.UpdateValue)
				value.DELETE("/:id", h.EnumValueHandler.DeleteValue)
			}
		}
	}

	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
