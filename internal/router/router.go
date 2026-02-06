package router

import (
	"os"

	"go-otel-example/internal/handlers"
	"go-otel-example/internal/health"
	"go-otel-example/internal/metrics"
	"go-otel-example/internal/observability"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func New(tracer trace.Tracer, meter metric.Meter, healthChecker *health.Checker) (*gin.Engine, error) {
	r := gin.Default()

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = observability.ServiceName()
	}

	r.Use(otelgin.Middleware(serviceName))

	httpMetrics, err := metrics.NewHTTPMetrics(meter, metrics.PrefixFromEnv())
	if err != nil {
		return nil, err
	}
	r.Use(httpMetrics.Middleware())

	handler := handlers.New(tracer, healthChecker, serviceName)

	r.GET("/health", handler.Health)
	r.GET("/test", handler.Test)
	r.GET("/test/:id", handler.TestWithID)
	r.POST("/log", handler.Log)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r, nil
}
