// Package routers регистрирует все хендлеры
package routers

import (
	"Brewery/internal/http/handlers"
	"Brewery/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRouters(e *gin.Engine, h handlers.Handlers) {
	api := e.Group("/api")
	admin := api.Group("", middleware.AdminAuth())

	registerBeerRouters(api, admin, h)
	registerReviewRouters(api, admin, h)
	registerCategoryRouters(api, admin, h)
	registerEnumRouters(api, h)
	registerAggregatesRouters(api, h)

	api.POST("/login", h.AuthHandler.Login)
	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func registerBeerRouters(api *gin.RouterGroup, admin *gin.RouterGroup, h handlers.Handlers) {
	beers := api.Group("/beers")
	{
		beers.GET("", h.BeersHandler.GetAllBeers)
		beers.GET("/search", h.BeersHandler.SearchBeer)

		features := beers.Group("/feats")
		{
			features.GET("/:beer_id", h.BeersHandler.GetFeature)
		}
	}

	adminBeers := admin.Group("/beers")
	{
		adminBeers.POST("", h.BeersHandler.CreateBeer)
		adminBeers.PATCH("/:id", h.BeersHandler.UpdateBeer)
		adminBeers.DELETE("/:id", h.BeersHandler.DeleteBeer)

		features := adminBeers.Group("/feats")
		{
			features.POST("/:beer_id", h.BeersHandler.CreateFeature)
			features.PATCH("/:beer_id", h.BeersHandler.UpdateBeer)
			features.DELETE("/:beer_id", h.BeersHandler.DeleteFeature)
		}
	}
}

func registerReviewRouters(api *gin.RouterGroup, admin *gin.RouterGroup, h handlers.Handlers) {
	reviews := api.Group("/reviews")
	{
		reviews.POST("/:beer_id", h.ReviewHandler.CreateReview)
		reviews.GET("/:beer_id", h.ReviewHandler.GetBeersReviews)
	}

	adminReviews := admin.Group("/reviews")
	{
		adminReviews.DELETE("/:id", h.ReviewHandler.DeleteReview)
		adminReviews.PATCH("/:id", h.ReviewHandler.UpdateReview)
	}
}

func registerCategoryRouters(api *gin.RouterGroup, admin *gin.RouterGroup, h handlers.Handlers) {
	categories := api.Group("/categories")
	{
		categories.GET("/:id", h.CategoryHandler.GetCategoryByID)
		categories.GET("", h.CategoryHandler.GetAllCategories)
		categories.GET("/beers/:category_id", h.CategoryHandler.GetBeersByCategory)
		categories.GET("/parent/:id", h.CategoryHandler.GetParentCategory)
		categories.GET("/children/:id", h.CategoryHandler.GetChildCategory)
		categories.GET("/:id/beers/search", h.BeersHandler.SearchBeer)

		parameters := categories.Group("/parameters")
		{
			parameters.GET("", h.ParametersHandler.ListCategoryParameters)
		}
	}

	adminCategories := admin.Group("/categories")
	{
		adminCategories.POST("", h.CategoryHandler.CreateCategory)
		adminCategories.PATCH("/:id", h.CategoryHandler.UpdateCategory)
		adminCategories.DELETE("/:id", h.CategoryHandler.DeleteCategory)

		adminParameters := categories.Group("/parameters")
		{
			adminParameters.PATCH("/apply/:category_id", h.ParametersHandler.ApplyParametersToCategory)
			adminParameters.POST("/numeric", h.ParametersHandler.CreateNumericParameter)
			adminParameters.PATCH("/:id", h.ParametersHandler.UpdateParameter)
			adminParameters.DELETE("/:id", h.ParametersHandler.DeleteParameter)
			adminParameters.POST("/enum", h.ParametersHandler.CreateEnumParameter)
		}
	}
}

func registerEnumRouters(
	admin *gin.RouterGroup,
	h handlers.Handlers,
) {
	enums := admin.Group("/enums")
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

func registerAggregatesRouters(
	admin *gin.RouterGroup,
	h handlers.Handlers,
) {
	aggregates := admin.Group("/aggregates")
	{
		aggregates.PATCH("/apply/:category_id", h.AggregatesHandler.ApplyAggregateToCategory)
		aggregates.PATCH("/:id", h.AggregatesHandler.UpdateAggregate)
		aggregates.DELETE("/:id", h.AggregatesHandler.DeleteAggregate)
		aggregates.POST("", h.AggregatesHandler.CreateAggregate)
		aggregates.GET("", h.AggregatesHandler.GetAggregates)
	}
}
