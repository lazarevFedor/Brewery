// Package main содержит функцию main(), которая является точкой входа в программу.
package main

import (
	"Brewery/internal/config"
	"Brewery/internal/http/handlers"
	"Brewery/internal/http/middleware"
	"Brewery/internal/repository"
	"Brewery/internal/usecase"
	"Brewery/pkg/logger"
	"Brewery/pkg/postgres"
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cors "github.com/rs/cors/wrapper/gin"
)


const devMode = true

// main является точкой входа в программу.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, err := logger.NewLoggerContext(ctx, devMode)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	cfg := config.NewAppConfig()
	if cfg == nil {
		panic(errors.New("failed to create config"))
	}

	pool, err := postgres.NewPool(ctx, cfg.Postgres)
	if err != nil {
		panic(fmt.Errorf("failed to create postgres pool: %w", err))
	}

	beerRepo := repository.NewBeerPostgres(pool)
	ctgRepo := repository.NewCategoryPostgres(pool)

	beerSrv := usecase.NewBeerService(beerRepo, ctgRepo)
	beersHandler := handlers.NewBeersHandlers(beerSrv)
	categoryHandler := handlers.NewCategoriesHandlers(beerSrv)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(func(c *gin.Context) {
		ctxWithLogger := logger.NewContextWithLogger(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctxWithLogger)
		c.Next()
	})
	engine.Use(middleware.RequestContextMiddleware())
	engine.Use(middleware.MetricsMiddleware())
	engine.Use(cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))

	router, err := graceful.New(
		engine,
		graceful.WithAddr(fmt.Sprintf(":%s", cfg.Port)),
		graceful.WithShutdownTimeout(graceful.DefaultShutdownTimeout),
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize router: %w", err))
	}
	defer router.Close()

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.POST("/api/categories", categoryHandler.CreateCategory)
	router.GET("/api/categories/:id", categoryHandler.GetCategoryById)
	router.PATCH("/api/categories/:id", categoryHandler.UpdateCategory)
	router.DELETE("/api/categories/:id", categoryHandler.DeleteCategory)
	router.GET("/api/categories", categoryHandler.GetAllCategories)
	router.GET("/api/categories/beers/:category_id", categoryHandler.GetBeersByCategory)
	router.GET("/api/categories/parent/:id", categoryHandler.GetParentCategory)
	router.GET("/api/categories/children/:id", categoryHandler.GetChildCategory)

	router.POST("/api/beers", beersHandler.CreateBeer)
	router.PATCH("/api/beers/:id", beersHandler.UpdateBeer)
	router.DELETE("/api/beers/:id", beersHandler.DeleteBeer)
	router.GET("/api/beers", beersHandler.GetAllBeers)
	router.POST("/api/beers/reviews/:beer_id", beersHandler.CreateBeerReview)

	log.Info(ctx, fmt.Sprintf("server listening on port %s", cfg.Port))

	if err = router.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("failed to run router: %w", err))
	}
}
