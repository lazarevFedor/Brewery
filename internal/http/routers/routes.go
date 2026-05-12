package routers

import (
	"Brewery/internal/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(e *gin.Engine, h handlers.Handlers) {
	api := e.Group("/api")

	registerBeerRoutes(api, h)
	registerReviewRoutes(api, h)
	registerCategoryRoutes(api, h)
	registerEnumRoutes(api, h)
	registerAggregatesRoutes(api, h)

	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func registerBeerRoutes(api *gin.RouterGroup, h handlers.Handlers) {
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
}

func registerReviewRoutes(api *gin.RouterGroup, h handlers.Handlers) {
	reviews := api.Group("/reviews")
	{
		reviews.POST("/:beer_id", h.ReviewHandler.CreateReview)
		reviews.GET("/:beer_id", h.ReviewHandler.GetBeersReviews)
		reviews.DELETE("/:id", h.ReviewHandler.DeleteReview)
		reviews.PATCH("/:id", h.ReviewHandler.UpdateReview)
	}
}

func registerCategoryRoutes(api *gin.RouterGroup, h handlers.Handlers) {
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

		params := categories.Group("/parameters")
		{
			params.GET("", h.ParametersHandler.ListCategoryParameters)
			params.PATCH("/:category_id/apply", h.ParametersHandler.ApplyParametersToCategory)
			params.POST("/numeric", h.ParametersHandler.CreateNumericParameter)
			params.PATCH("/numeric/:id", h.ParametersHandler.UpdateNumericParameter)
			params.DELETE("/numeric/:id", h.ParametersHandler.DeleteNumericParameter)
			params.POST("/enum", h.ParametersHandler.CreateEnumParameter)
			params.PATCH("/enum/:id", h.ParametersHandler.UpdateEnumParameter)
			params.DELETE("/enum/:id", h.ParametersHandler.DeleteEnumParameter)
		}
	}
}

func registerEnumRoutes(api *gin.RouterGroup, h handlers.Handlers) {
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
}

func registerAggregatesRoutes(api *gin.RouterGroup, h handlers.Handlers) {
	aggregates := api.Group("/aggregates")
	{
		aggregates.PATCH("/apply/:category_id", h.AggregatesHandler.ApplyAggregateToCategory)
		aggregates.PATCH("/:id", h.AggregatesHandler.UpdateAggregate)
		aggregates.DELETE("/:id", h.AggregatesHandler.DeleteAggregate)
		aggregates.POST("", h.AggregatesHandler.CreateAggregate)
		aggregates.GET("", h.AggregatesHandler.GetAggregates)

		value := aggregates.Group("/value")
		{
			value.PATCH("/apply/:category_id", h.AggregatesHandler.ApplyAggregateToCategory)
			value.PATCH("/:id", h.AggregatesHandler.UpdateAggregate)
			value.DELETE("/:id", h.AggregatesHandler.DeleteAggregate)
			value.POST("", h.AggregatesHandler.CreateAggregate)
			value.GET("", h.AggregatesHandler.GetAggregates)

		}
	}
}
