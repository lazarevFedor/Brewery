package middleware

import (
	"Brewery/pkg/logger"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	timings = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Request latency in seconds by method, route and status",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method, route and status",
		},
		[]string{"method", "route", "status"},
	)
	metricsRegisterOnce sync.Once
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

func MetricsMiddleware() gin.HandlerFunc {
	metricsRegisterOnce.Do(func() {
		prometheus.MustRegister(timings, counter)
	})

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		status := strconv.Itoa(c.Writer.Status())

		timings.
			WithLabelValues(c.Request.Method, route, status).
			Observe(time.Since(start).Seconds())
		counter.
			WithLabelValues(c.Request.Method, route, status).
			Inc()
	}
}
