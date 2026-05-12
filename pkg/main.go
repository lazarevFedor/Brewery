package main

import (
	"Brewery/pkg/logger"
	"context"
	"log"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	ctx, err := logger.NewLoggerContext(ctx, true)
	if err != nil {
		log.Fatal(err)
	}

	logg, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		log.Fatal("logger not found in context")
	}

	logg.Debug(ctx, "Я чмо")
	logg.Info(ctx, "Писичка", zap.String("Кто канибал", "Я"))
}
