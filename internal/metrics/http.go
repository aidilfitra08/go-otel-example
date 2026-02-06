package metrics

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const defaultMetricsPrefix = "example_app_"

func PrefixFromEnv() string {
	prefix := os.Getenv("METRICS_PREFIX")
	if prefix == "" {
		return defaultMetricsPrefix
	}

	return prefix
}

type HTTPMetrics struct {
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	requestSize     metric.Int64Histogram
	responseSize    metric.Int64Histogram
}

func NewHTTPMetrics(meter metric.Meter, prefix string) (*HTTPMetrics, error) {
	requestCounter, err := meter.Int64Counter(prefix+"http.requests.total", metric.WithDescription("Total HTTP requests"))
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram(prefix+"http.request.duration", metric.WithDescription("HTTP request duration in seconds"))
	if err != nil {
		return nil, err
	}

	requestSize, err := meter.Int64Histogram(prefix+"http.request.size", metric.WithDescription("HTTP request size in bytes"))
	if err != nil {
		return nil, err
	}

	responseSize, err := meter.Int64Histogram(prefix+"http.response.size", metric.WithDescription("HTTP response size in bytes"))
	if err != nil {
		return nil, err
	}

	return &HTTPMetrics{
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
		requestSize:     requestSize,
		responseSize:    responseSize,
	}, nil
}

func (m *HTTPMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		if c.Request.ContentLength > 0 {
			m.requestSize.Record(c.Request.Context(), c.Request.ContentLength, metric.WithAttributes(
				attribute.String("method", method),
				attribute.String("path", path),
			))
		}

		c.Next()

		duration := time.Since(start).Seconds()
		attrs := metric.WithAttributes(
			attribute.String("method", method),
			attribute.String("path", path),
			attribute.Int("status", c.Writer.Status()),
		)

		m.requestCounter.Add(c.Request.Context(), 1, attrs)
		m.requestDuration.Record(c.Request.Context(), duration, attrs)
		m.responseSize.Record(c.Request.Context(), int64(c.Writer.Size()), attrs)
	}
}
