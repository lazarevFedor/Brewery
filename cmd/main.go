package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
)

func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router, err := graceful.Default()
	if err != nil {
		panic(err)
	}
	defer router.Close()

	router.Use(RequestContextMiddleware())

	if err = router.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}
