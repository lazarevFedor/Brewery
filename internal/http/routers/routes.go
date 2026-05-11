package routers

import (
	"Brewery/internal/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(
	e *gin.Engine,
	beersHandler handlers.BeersHandlers,
	reviewHandler handlers.ReviewsHandlers,
	categoryHandler handlers.CategoriesHandlers,
	enumHandler handlers.EnumHandlers,
	parametersHandler handlers.ParametersHandlers,
	aggregateHandler handlers.AggregateHandlers,
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

			parameters := categories.Group("/parameters")
			{
				parameters.GET("", parametersHandler.ListCategoryParameters)
				parameters.PATCH("/:category_id/apply", parametersHandler.ApplyParametersToCategory)

				parameters.POST("/numeric", parametersHandler.CreateNumericParameter)
				parameters.PATCH("/numeric/:id", parametersHandler.UpdateNumericParameter)
				parameters.DELETE("/numeric/:id", parametersHandler.DeleteNumericParameter)

				parameters.POST("/enum", parametersHandler.CreateEnumParameter)
				parameters.PATCH("/enum/:id", parametersHandler.UpdateEnumParameter)
				parameters.DELETE("/enum/:id", parametersHandler.DeleteEnumParameter)
			}
		}

		enums := api.Group("/enums")
		{
			enums.POST("", enumHandler.CreateEnum)
			enums.GET("", enumHandler.GetEnum)
			enums.PATCH("/:id", enumHandler.UpdateEnum)
			enums.DELETE("/:id", enumHandler.DeleteEnum)

			value := enums.Group("/value")
			{
				value.POST("", enumHandler.CreateValue)
				value.GET("", enumHandler.GetValue)
				value.PATCH("/:id", enumHandler.UpdateValue)
				value.DELETE("/:id", enumHandler.DeleteValue)
			}
		}

		aggregates := api.Group("/aggregates")
		{
			aggregates.POST("", aggregateHandler.CreateAggregate)
			aggregates.GET("/:id", aggregateHandler.GetAggregates)
			aggregates.PATCH("/:id", aggregateHandler.UpdateAggregate)
			aggregates.DELETE("/:id", aggregateHandler.DeleteAggregate)
			aggregates.PATCH("/:category_id/apply", aggregateHandler.ApplyAggregateToCategory)
		}
	}

	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
