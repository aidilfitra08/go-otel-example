package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
)

func (h *Handler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	_, span := h.tracer.Start(ctx, "health-check")
	defer span.End()

	overall, dependencies := h.health.Snapshot()
	statusText := "healthy"
	statusCode := http.StatusOK
	if !overall {
		statusText = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	span.SetAttributes(
		attribute.String("health.status", statusText),
		attribute.String("component", "health-handler"),
	)

	response := gin.H{
		"status":  statusText,
		"service": h.serviceName,
		"time":    time.Now().Format(time.RFC3339),
	}

	if parseBoolQuery(c.Query("details")) {
		deps := make(map[string]string, len(dependencies))
		for name, healthy := range dependencies {
			if healthy {
				deps[name] = "healthy"
			} else {
				deps[name] = "unhealthy"
			}
		}
		response["dependencies"] = deps
	}

	c.JSON(statusCode, response)
}
