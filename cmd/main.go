package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joaziz/go-gin-graceful-shutdown/graceful"
)

func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func main() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(RequestContextMiddleware())
	graceful.New(router, 30*time.Second).ListenAndServe(":8080")
}
