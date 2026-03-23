package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "clamav_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"endpoint", "method", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "clamav_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"endpoint", "method"},
	)

	filesScannedClean = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "clamav_files_scanned_clean_total",
			Help: "Total number of clean files scanned",
		},
	)

	filesScannedInfected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "clamav_files_scanned_infected_total",
			Help: "Total number of infected files scanned",
		},
	)

	filesScannedError = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "clamav_files_scanned_error_total",
			Help: "Total number of scan errors",
		},
	)

	filesScannedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "clamav_files_scanned_total",
			Help: "Total number of files scanned",
		},
		[]string{"result"},
	)

	fileSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "clamav_file_size_bytes",
			Help:    "Size of uploaded files in bytes",
			Buckets: []float64{1024, 10240, 102400, 1048576, 10485760, 104857600},
		},
		[]string{"operation"},
	)
)

func PrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()
		path := c.Route().Path

		httpRequestsTotal.WithLabelValues(path, method, status).Inc()
		httpRequestDuration.WithLabelValues(path, method).Observe(duration)

		return err
	}
}

func RecordFileScanned(result string) {
	filesScannedTotal.WithLabelValues(result).Inc()

	switch result {
	case "clean":
		filesScannedClean.Inc()
	case "infected":
		filesScannedInfected.Inc()
	case "error":
		filesScannedError.Inc()
	}
}

func RecordFileSize(operation string, size int64) {
	fileSizeBytes.WithLabelValues(operation).Observe(float64(size))
}
