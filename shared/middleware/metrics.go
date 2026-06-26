package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of in-flight HTTP requests",
	})

	rateLimitExceededTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rate_limit_exceeded_total",
		Help: "Total number of rate limited requests",
	})
)

func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		start := time.Now()
		path := c.Path()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		httpRequestsTotal.WithLabelValues(c.Method(), path, strconv.Itoa(status)).Inc()
		httpRequestDuration.WithLabelValues(c.Method(), path).Observe(duration.Seconds())

		return err
	}
}

func MetricsHandler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}

func RecordRateLimitExceeded() {
	rateLimitExceededTotal.Inc()
}
