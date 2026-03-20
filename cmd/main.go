package main

import (
	"Brewery/internal/config"
	"Brewery/pkg/logger"
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := uuid.New().String()[:8]

		ctx := c.Request.Context()

		ctx = logger.WithRequestID(ctx, reqID)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

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

	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create config: %w", err))
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(RequestContextMiddleware())

	router, err := graceful.New(
		engine,
		graceful.WithAddr(fmt.Sprintf(":%s", cfg.Port)),
		graceful.WithShutdownTimeout(graceful.DefaultShutdownTimeout),
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize router: %w", err))
	}
	defer router.Close()

	log.Info(ctx, fmt.Sprintf("server listening on port %s", cfg.Port))
	if err = router.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("failed to run router: %w", err))
	}
}
