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

			parameters := categories.Group("/parameters")
			{
				parameters.GET("", h.ParametersHandler.ListCategoryParameters)
				parameters.PATCH("/:category_id/apply", h.ParametersHandler.ApplyParametersToCategory)

				parameters.POST("/numeric", h.ParametersHandler.CreateNumericParameter)
				parameters.PATCH("/numeric/:id", h.ParametersHandler.UpdateNumericParameter)
				parameters.DELETE("/numeric/:id", h.ParametersHandler.DeleteNumericParameter)

				parameters.POST("/enum", h.ParametersHandler.CreateEnumParameter)
				parameters.PATCH("/enum/:id", h.ParametersHandler.UpdateEnumParameter)
				parameters.DELETE("/enum/:id", h.ParametersHandler.DeleteEnumParameter)
			}
		}

		enums := api.Group("/enums")
		{
			enums.POST("", h.EnumHandler.CreateEnum)
			enums.GET("", h.EnumHandler.GetEnum)
			enums.PATCH("/:id", h.EnumHandler.UpdateEnum)
			enums.DELETE("/:id", h.EnumHandler.DeleteEnum)

			value := enums.Group("/value")
			{
				value.POST("", h.EnumHandler.CreateValue)
				value.GET("", h.EnumHandler.GetValue)
				value.PATCH("/:id", h.EnumHandler.UpdateValue)
				value.DELETE("/:id", h.EnumHandler.DeleteValue)
			}
		}

		aggregates := api.Group("/aggregates")
		{
			aggregates.GET("", h.AggregatesHandler.GetAggregates)
			aggregates.POST("", h.AggregatesHandler.CreateAggregate)
			aggregates.PATCH("/:id", h.AggregatesHandler.UpdateAggregate)
			aggregates.DELETE("/:id", h.AggregatesHandler.DeleteAggregate)
			aggregates.PATCH("/:category_id/apply", h.AggregatesHandler.ApplyAggregateToCategory)
		}
	}

	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
