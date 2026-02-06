package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handler) Test(c *gin.Context) {
	ctx := c.Request.Context()
	_, span := h.tracer.Start(ctx, "test-endpoint")
	defer span.End()

	time.Sleep(50 * time.Millisecond)

	span.SetAttributes(
		attribute.String("endpoint", "test"),
		attribute.Int("simulated.delay.ms", 50),
	)

	span.AddEvent("Processing test request", trace.WithAttributes(
		attribute.String("request.id", c.GetHeader("X-Request-ID")),
	))

	c.JSON(http.StatusOK, gin.H{
		"message":  "Test endpoint working!",
		"trace_id": span.SpanContext().TraceID().String(),
		"span_id":  span.SpanContext().SpanID().String(),
	})
}

func (h *Handler) TestWithID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	_, span := h.tracer.Start(ctx, "test-endpoint-with-id")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint", "test-with-id"),
		attribute.String("resource.id", id),
	)

	time.Sleep(100 * time.Millisecond)

	span.AddEvent("Processing resource", trace.WithAttributes(
		attribute.String("resource.id", id),
		attribute.String("action", "fetch"),
	))

	c.JSON(http.StatusOK, gin.H{
		"message":  "Test endpoint with ID",
		"id":       id,
		"trace_id": span.SpanContext().TraceID().String(),
	})
}
