package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"go-otel-example/internal/health"
	"go-otel-example/internal/metrics"
	"go-otel-example/internal/observability"
	"go-otel-example/internal/router"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()

	// Initialize tracer
	tracer, tp, err := observability.InitTracer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Initialize metrics
	meter, mp, err := observability.InitMeterProvider(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mp.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}()

	healthChecker, err := health.NewChecker(meter, metrics.PrefixFromEnv())
	if err != nil {
		log.Fatalf("Failed to initialize health checks: %v", err)
	}
	healthChecker.SetStatus("db", true)
	healthChecker.SetStatus("redis", true)

	r, err := router.New(tracer, meter, healthChecker)
	if err != nil {
		log.Fatalf("Failed to initialize router: %v", err)
	}

	// Start server in a goroutine
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Server starting on port %s", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")
}
