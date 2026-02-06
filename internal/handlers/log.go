package handlers

import (
	"net/http"

	"go-otel-example/internal/logging"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
)

func (h *Handler) Log(c *gin.Context) {
	ctx := c.Request.Context()
	_, span := h.tracer.Start(ctx, "log-endpoint")
	defer span.End()

	var body struct {
		Message string `json:"message" binding:"required"`
		Level   string `json:"level"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	level := body.Level
	if level == "" {
		level = "info"
	}

	if err := logging.SendToLoki(h.serviceName, level, body.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send log"})
		return
	}

	span.SetAttributes(
		attribute.String("log.message", body.Message),
		attribute.String("log.level", level),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "logged",
		"message": body.Message,
		"level":   level,
	})
}
