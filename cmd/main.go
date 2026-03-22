package main

import (
	"Brewery/internal/config"
	"Brewery/internal/http/handlers"
	"Brewery/internal/http/middleware"
	repository "Brewery/internal/repository/beer"
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
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, err := logger.NewLoggerContext(ctx, true)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	cfg, err := config.FillConfig(config.NewAppConfig())
	if err != nil {
		panic(fmt.Errorf("failed to create config: %w", err))
	}

	pool, err := postgres.NewPool(ctx, cfg.Postgres)
	if err != nil {
		panic(fmt.Errorf("failed to create postgres pool: %w", err))
	}

	beerRepo := repository.NewBeerPostgres(pool)

	// TODO: add category repository
	var ctgRepo any

	beerSrv := usecase.NewBeerService(beerRepo, ctgRepo)
	_ = handlers.NewBeersHandlers(beerSrv)
	_ = handlers.NewCategoriesHandlers(beerSrv)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestContextMiddleware())
	engine.Use(middleware.MetricsMiddleware())

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

	log.Info(ctx, fmt.Sprintf("server listening on port %s", cfg.Port))

	if err = router.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("failed to run router: %w", err))
	}
}
