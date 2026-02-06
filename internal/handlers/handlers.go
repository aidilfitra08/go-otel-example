package handlers

import (
	"strings"

	"go-otel-example/internal/health"

	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	tracer      trace.Tracer
	health      *health.Checker
	serviceName string
}

func New(tracer trace.Tracer, healthChecker *health.Checker, serviceName string) *Handler {
	return &Handler{
		tracer:      tracer,
		health:      healthChecker,
		serviceName: serviceName,
	}
}

func parseBoolQuery(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "true" || value == "1" || value == "yes" || value == "y" || value == "on"
}
