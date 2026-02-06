# Go Gin Server with OpenTelemetry

A complete example of a Go Gin web server instrumented with OpenTelemetry, exporting traces and logs to Tempo and Loki, with visualization in Grafana. The setup uses two separate Docker networks to demonstrate a production-like architecture.

## Architecture

```
┌─────────────────────┐
│   Go Gin Server     │
│  (OpenTelemetry)    │
└──────────┬──────────┘
           │
           │ OTLP (gRPC)
           ▼
┌─────────────────────┐      ┌────────────────────────┐
│  OTel Collector     │◄────►│  Observability Stack   │
│  (otel-network)     │      │ (observability-network)│
└─────────────────────┘      │                        │
                             │  - Tempo (traces)      │
                             │  - Loki (logs)         │
                             │  - Grafana (UI)        │
                             └────────────────────────┘
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Windows (PowerShell), Linux, or macOS

## Project Structure

```
.
├── main.go                           # Go Gin server with OTel middleware
├── go.mod                            # Go dependencies
├── docker-compose-observability.yaml # Tempo, Loki, Grafana stack
├── docker-compose-otel.yaml          # OpenTelemetry Collector
├── otel-collector-config.yaml        # OTel Collector configuration
├── tempo-config.yaml                 # Tempo configuration
├── loki-config.yaml                  # Loki configuration
├── grafana-datasources.yaml          # Grafana datasources
├── grafana-dashboards.yaml           # Grafana dashboard provisioning
└── grafana-dashboard-otel.json       # Pre-built Grafana dashboard
```

## Quick Start

### Step 1: Start the Observability Stack

First, start Tempo, Loki, and Grafana:

```powershell
docker-compose -f docker-compose-observability.yaml up -d
```

This creates the `observability-network` Docker network and starts:

- **Tempo** on port 3200 (for traces)
- **Loki** on port 3100 (for logs)
- **Grafana** on port 3000 (for visualization)

### Step 2: Start the OpenTelemetry Collector

Next, start the OTel Collector:

```powershell
docker-compose -f docker-compose-otel.yaml up -d
```

The OTel Collector:

- Connects to both `otel-network` and `observability-network`
- Receives traces/logs on port 4317 (gRPC) and 4318 (HTTP)
- Forwards data to Tempo and Loki

### Step 3: Run the Go Server

Install dependencies and run the server:

```powershell
# Install Go dependencies
go mod download

# Run the server
$env:OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4317"
$env:PORT="8080"
go run main.go
```

The server will start on port 8080 with the following endpoints:

- `GET /health` - Health check endpoint
- `GET /test` - Test endpoint with tracing
- `GET /test/:id` - Test endpoint with parameter

### Step 4: Access Grafana

Open Grafana in your browser:

```
http://localhost:3000
```

The dashboard "Go Gin Server - OpenTelemetry" will be automatically provisioned and available.

## Testing the Setup

Generate some traffic to see traces and logs:

```powershell
# Health check
curl http://localhost:8080/health

# Test endpoint
curl http://localhost:8080/test

# Test with ID
curl http://localhost:8080/test/123

# Generate multiple requests
1..10 | ForEach-Object { curl http://localhost:8080/test }
```

## Viewing Observability Data

### Grafana Dashboard

The pre-built dashboard shows:

- **Total Traces**: Number of traces received
- **Average Response Time**: Mean latency across all endpoints
- **Log Rate**: Rate of logs being ingested
- **Trace Search**: Browse and search traces
- **Logs**: View application logs
- **Request Rate by Endpoint**: Traffic distribution
- **P95 Latency by Endpoint**: 95th percentile latency per endpoint

### Manual Exploration

1. **Explore Traces in Tempo**:
   - Go to Grafana → Explore
   - Select "Tempo" datasource
   - Use TraceQL: `{service.name="go-gin-server"}`

2. **Explore Logs in Loki**:
   - Go to Grafana → Explore
   - Select "Loki" datasource
   - Use LogQL: `{job="otel-collector"}`

3. **Link Traces to Logs**:
   - Click on any trace span
   - Click "Logs for this span" to see correlated logs

## Configuration Details

### OpenTelemetry Collector

The collector receives telemetry data and exports it to multiple backends:

- **Receivers**: OTLP (gRPC on 4317, HTTP on 4318)
- **Processors**: Batch, Memory Limiter, Resource
- **Exporters**:
  - Tempo (traces via OTLP)
  - Loki (logs via HTTP)
  - Prometheus (metrics on port 8889)
  - Logging (console, for debugging)

### Network Architecture

Two separate Docker networks:

1. **otel-network**: For the OTel Collector
2. **observability-network**: For Tempo, Loki, and Grafana

The OTel Collector bridges both networks, allowing:

- The Go server to send data to the collector on `otel-network`
- The collector to forward data to backends on `observability-network`

## Environment Variables

### Go Server

- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTel Collector endpoint (default: `localhost:4317`)
- `PORT`: Server port (default: `8080`)

## Stopping the Services

Stop all services:

```powershell
# Stop the Go server (Ctrl+C)

# Stop the OTel Collector
docker-compose -f docker-compose-otel.yaml down

# Stop the observability stack
docker-compose -f docker-compose-observability.yaml down
```

To remove volumes and clean up data:

```powershell
docker-compose -f docker-compose-otel.yaml down -v
docker-compose -f docker-compose-observability.yaml down -v
```

## Troubleshooting

### OTel Collector not connecting to backends

Check that the observability stack is running:

```powershell
docker-compose -f docker-compose-observability.yaml ps
```

View OTel Collector logs:

```powershell
docker logs otel-collector
```

### Go server cannot connect to OTel Collector

Verify the collector is running and accessible:

```powershell
curl http://localhost:13133/
```

### No data in Grafana

1. Check OTel Collector is receiving data: `http://localhost:8888/metrics`
2. Verify Tempo is receiving traces: `http://localhost:3200/status`
3. Check Loki is receiving logs: `http://localhost:3100/ready`

### Network issues

Ensure the observability network exists:

```powershell
docker network ls | Select-String observability-network
```

If not, start the observability stack first.

## Customization

### Adding More Endpoints

Edit `main.go` to add new endpoints with automatic tracing:

```go
r.GET("/custom", func(c *gin.Context) {
    ctx := c.Request.Context()
    _, span := tracer.Start(ctx, "custom-endpoint")
    defer span.End()

    // Your logic here

    c.JSON(200, gin.H{"message": "Custom endpoint"})
})
```

### Modifying the Dashboard

1. Edit the dashboard in Grafana UI
2. Export the dashboard JSON
3. Save it to `grafana-dashboard-otel.json`
4. Restart Grafana to apply changes

### Adding More Exporters

Edit `otel-collector-config.yaml` to add exporters like Jaeger, Zipkin, or Prometheus RemoteWrite.

## Learn More

- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Grafana Tempo](https://grafana.com/docs/tempo/latest/)
- [Grafana Loki](https://grafana.com/docs/loki/latest/)
- [Gin Web Framework](https://gin-gonic.com/)

## License

MIT License - feel free to use this as a template for your projects!
