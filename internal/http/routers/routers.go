// Package routers регистрирует все хендлеры
package routers

import (
	"Brewery/internal/http/handlers"
	"Brewery/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterRouters регистрирует маршруты API в соответствии с api/services.openapi.json.
// Запросы, помеченные в swagger как bearerAuth, регистрируются в группе admin.
func RegisterRouters(e *gin.Engine, h handlers.Handlers) {
	api := e.Group("/api")
	admin := api.Group("", middleware.AdminAuth())

	api.POST("/login", h.AuthHandler.Login)

	registerClassRouters(api, admin, h)
	registerProductRouters(api, admin, h)
	registerMaterialRouters(api, admin, h)
	registerHistoryRouters(api, admin, h)
	registerEnumParameterRouters(admin, h)
	registerEnumValueRouters(admin, h)
	registerRangeParameterRouters(admin, h)
	registerAggregateRouters(admin, h)
	registerEquipmentRouters(api, admin, h)
	registerOperationRouters(api, admin, h)

	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func registerClassRouters(api, admin *gin.RouterGroup, h handlers.Handlers) {
	api.GET("/classes", h.ClassHandler.ListClasses)
	api.GET("/classes/:id", h.ClassHandler.GetClass)
	api.GET("/classes/parent/:id", h.ClassHandler.GetClassParent)
	api.GET("/classes/children/:id", h.ClassHandler.ListClassChildren)

	admin.POST("/classes", h.ClassHandler.CreateClass)
	admin.PATCH("/classes/:id", h.ClassHandler.UpdateClass)
	admin.DELETE("/classes/:id", h.ClassHandler.DeleteClass)

	api.GET("/classes/parameters", h.ClassHandler.ListParameters)
	admin.PATCH("/classes/parameters/apply/:id", h.ClassHandler.ApplyParametersToClass)
	admin.PATCH("/classes/parameters/exclude/:id", h.ClassHandler.ExcludeParametersFromClass)
}

func registerProductRouters(api, admin *gin.RouterGroup, h handlers.Handlers) {
	api.GET("/products", h.ProductHandler.ListProducts)
	api.GET("/products/search", h.ProductHandler.SearchProducts)
	api.GET("/products/:id", h.ProductHandler.GetProduct)
	api.GET("/products/:id/derived", h.ProductHandler.ListDerivedProducts)
	api.GET("/classes/products/:class_id", h.ProductHandler.ListProductsByClass)
	api.GET("/classes/:id/products/search", h.ProductHandler.SearchProductsByClass)

	admin.POST("/products", h.ProductHandler.CreateProduct)
	admin.PATCH("/products/:id", h.ProductHandler.UpdateProduct)
	admin.DELETE("/products/:id", h.ProductHandler.DeleteProduct)
}

func registerMaterialRouters(api, admin *gin.RouterGroup, h handlers.Handlers) {
	api.GET("/products/:id/materials", h.MaterialHandler.GetProductBomTree)
	api.GET("/products/:id/materials/norms", h.MaterialHandler.CalculateMaterialNorms)
	api.GET("/materials/:material_id/products", h.MaterialHandler.ListProductsByMaterial)

	admin.POST("/products/:id/materials", h.MaterialHandler.AddProductMaterial)
	admin.PATCH("/products/:id/materials/:material_id", h.MaterialHandler.UpdateProductMaterial)
	admin.DELETE("/products/:id/materials/:material_id", h.MaterialHandler.DeleteProductMaterial)
}

func registerHistoryRouters(api, admin *gin.RouterGroup, h handlers.Handlers) {
	api.GET("/products/:id/history", h.HistoryHandler.ListProductHistory)
	api.GET("/history/:id", h.HistoryHandler.GetHistoryRecord)

	admin.POST("/products/:id/history", h.HistoryHandler.RegisterProductChange)
}

// registerEnumParameterRouters регистрирует операции EnumParameters; все требуют авторизации.
func registerEnumParameterRouters(admin *gin.RouterGroup, h handlers.Handlers) {
	admin.POST("/enums", h.EnumParametersHandler.CreateEnumParameter)
	admin.GET("/enums", h.EnumParametersHandler.ListEnumParameters)
	admin.PATCH("/enums/:id", h.EnumParametersHandler.UpdateEnumParameter)
	admin.DELETE("/enums/:id", h.EnumParametersHandler.DeleteEnumParameter)
}

// registerEnumValueRouters регистрирует операции EnumsValue; все требуют авторизации.
func registerEnumValueRouters(admin *gin.RouterGroup, h handlers.Handlers) {
	admin.POST("/enums/value", h.EnumValuesHandler.CreateEnumValue)
	admin.GET("/enums/value", h.EnumValuesHandler.ListEnumValues)
	admin.PATCH("/enums/value/:id", h.EnumValuesHandler.UpdateEnumValue)
	admin.DELETE("/enums/value/:id", h.EnumValuesHandler.DeleteEnumValue)
}

func registerRangeParameterRouters(admin *gin.RouterGroup, h handlers.Handlers) {
	admin.POST("/range", h.RangeParametersHandler.CreateRangeParameter)
	admin.GET("/range", h.RangeParametersHandler.ListRangeParameters)
	admin.PATCH("/range/:id", h.RangeParametersHandler.UpdateRangeParameter)
	admin.DELETE("/range/:id", h.RangeParametersHandler.DeleteRangeParameter)
}

// registerAggregateRouters - по swagger все запросы агрегатов требуют авторизации, включая GET.
func registerAggregateRouters(admin *gin.RouterGroup, h handlers.Handlers) {
	admin.POST("/aggregates", h.AggregatesHandler.CreateAggregate)
	admin.GET("/aggregates", h.AggregatesHandler.GetAggregates)
	admin.PATCH("/aggregates/apply/:class_id", h.AggregatesHandler.ApplyAggregateToClass)
	admin.PATCH("/aggregates/:id", h.AggregatesHandler.UpdateAggregate)
	admin.DELETE("/aggregates/:id", h.AggregatesHandler.DeleteAggregate)
}

func registerEquipmentRouters(api, admin *gin.RouterGroup, h handlers.Handlers) {
	if h.EquipmentHandler == nil {
		return
	}
	api.GET("/equipment", h.EquipmentHandler.ListEquipment)
	api.GET("/equipment/:name/:work_center", h.EquipmentHandler.GetEquipment)
	api.GET("/equipment/:name", h.EquipmentHandler.GetEquipment)

	admin.POST("/equipment", h.EquipmentHandler.CreateEquipment)
	admin.PATCH("/equipment/:name/:work_center", h.EquipmentHandler.UpdateEquipment)
	admin.PATCH("/equipment/:name", h.EquipmentHandler.UpdateEquipment)
	admin.DELETE("/equipment/:name/:work_center", h.EquipmentHandler.DeleteEquipment)
	admin.DELETE("/equipment/:name", h.EquipmentHandler.DeleteEquipment)
}

func registerOperationRouters(api, admin *gin.RouterGroup, h handlers.Handlers) {
	if h.OperationHandler == nil {
		return
	}
	api.GET("/operations", h.OperationHandler.ListOperations)
	api.GET("/operations/:id", h.OperationHandler.GetOperation)
	api.GET("/products/:id/operations", h.OperationHandler.GetProductTechCard)

	admin.POST("/operations", h.OperationHandler.CreateOperation)
	admin.PATCH("/operations/:id", h.OperationHandler.UpdateOperation)
	admin.DELETE("/operations/:id", h.OperationHandler.DeleteOperation)
}
