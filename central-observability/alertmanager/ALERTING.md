# Alerting Setup Guide

This guide explains the alerting system implemented for your observability stack.

## Overview

The alerting system consists of:

- **Prometheus** - Evaluates alert rules and sends alerts
- **Alertmanager** - Routes and manages alert notifications
- **Alert Rules** - Defines conditions that trigger alerts

## Components Created

### 1. Alert Rules (`alert-rules.yml`)

Contains alert definitions organized by category:

#### Service Availability Alerts

- **ServiceDown**: Service is unreachable (Critical)
- **HighErrorRate**: Error rate > 5% for 5 minutes (Warning)
- **CriticalErrorRate**: Error rate > 20% for 2 minutes (Critical)

#### Performance Alerts

- **HighResponseTime**: P95 latency > 1 second (Warning)
- **VeryHighResponseTime**: P95 latency > 3 seconds (Critical)
- **HighRequestRate**: Request rate > 1000 req/s (Warning)

#### OTEL Collector Alerts

- **OTelCollectorDroppingSpans**: Collector is dropping spans (Warning)
- **OTelCollectorHighMemory**: Memory usage > 1GB (Warning)

#### Prometheus Alerts

- **PrometheusScrapeFailing**: Cannot scrape target (Warning)
- **PrometheusConfigReloadFailed**: Config reload failed (Warning)

#### Tempo Alerts

- **TempoIngestionFailing**: No blocks flushed in 10 minutes (Warning)

#### Application Specific Alerts

- **EndpointHighErrorRate**: Specific endpoint error rate > 10% (Warning)
- **SlowDatabaseQueries**: P95 query latency > 500ms (Warning)
- **HighActiveConnections**: Active connections > 100 (Warning)

### 2. Alertmanager Configuration (`alertmanager.yml`)

Manages alert routing and notifications:

- **Critical alerts**: Repeated every 1 hour, sent after 10 seconds
- **Warning alerts**: Repeated every 4 hours, sent after 30 seconds
- **Inhibition rules**: Suppress warning if critical alert exists

### 3. Updated Prometheus Configuration

- Alert evaluation interval: 30 seconds
- Connected to Alertmanager
- Loads alert rules from `alert-rules.yml`

## Getting Started

### 1. Start the Services

```powershell
cd central-observability
docker-compose up -d
```

### 2. Verify Services are Running

Check service status:

```powershell
docker-compose ps
```

You should see:

- prometheus-server (port 9090)
- alertmanager-server (port 9093)
- tempo-server, loki-server, grafana-server

### 3. Access the UIs

- **Prometheus**: http://localhost:9090
- **Alertmanager**: http://localhost:9093
- **Grafana**: http://localhost:3000

## Viewing Alerts

### In Prometheus

1. Go to http://localhost:9090/alerts
2. View all defined alert rules and their current states:
   - **Inactive** (green): Condition not met
   - **Pending** (yellow): Condition met, waiting for duration
   - **Firing** (red): Alert is active

### In Alertmanager

1. Go to http://localhost:9093
2. View active alerts being processed
3. See alert grouping and silencing options

### In Grafana

1. Go to http://localhost:3000
2. Navigate to **Alerting** → **Alert rules**
3. Or create dashboards with alert annotations

## Configuring Notifications

The default configuration logs alerts to the console. To send real notifications:

### Slack Integration

1. Create a Slack webhook URL
2. Edit `alertmanager.yml`:

```yaml
receivers:
  - name: "critical-receiver"
    slack_configs:
      - api_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
        channel: "#alerts-critical"
        title: "🚨 {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}"
        text: "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"
        send_resolved: true
```

3. Restart Alertmanager:

```powershell
docker-compose restart alertmanager
```

### Email Integration

Edit `alertmanager.yml`:

```yaml
receivers:
  - name: "critical-receiver"
    email_configs:
      - to: "oncall@example.com"
        from: "alertmanager@example.com"
        smarthost: "smtp.gmail.com:587"
        auth_username: "your-email@gmail.com"
        auth_password: "your-app-password"
        headers:
          Subject: "CRITICAL: {{ .GroupLabels.alertname }}"
```

### Discord Integration

Edit `alertmanager.yml`:

```yaml
receivers:
  - name: "critical-receiver"
    discord_configs:
      - webhook_url: "YOUR_DISCORD_WEBHOOK_URL"
        title: "🚨 CRITICAL ALERT"
        message: "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"
```

### Webhook Integration (Custom)

Edit `alertmanager.yml`:

```yaml
receivers:
  - name: "critical-receiver"
    webhook_configs:
      - url: "http://your-service/webhook/alerts"
        send_resolved: true
```

## Testing Alerts

### Test Service Down Alert

Stop a monitored service:

```powershell
# If you have the go-gin-server running
# Stop it to trigger ServiceDown alert
```

Wait 1 minute and check http://localhost:9090/alerts

### Test High Error Rate Alert

Generate errors in your application and monitor the alert status.

### Simulate Alert with amtool

```powershell
# Install amtool (Alertmanager CLI)
docker exec -it alertmanager-server amtool alert add test_alert severity=critical
```

## Customizing Alerts

### Adjust Alert Thresholds

Edit `alert-rules.yml` to modify thresholds:

```yaml
# Example: Change high error rate threshold from 5% to 10%
- alert: HighErrorRate
  expr: |
    (
      sum(rate(http_server_duration_count{http_status_code=~"5.."}[5m])) by (job, instance)
      /
      sum(rate(http_server_duration_count[5m])) by (job, instance)
    ) > 0.10  # Changed from 0.05 to 0.10
```

### Add Custom Alerts

Add new rules to `alert-rules.yml`:

```yaml
- name: custom_alerts
  interval: 30s
  rules:
    - alert: CustomMetricThreshold
      expr: your_custom_metric > 100
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Custom metric exceeded threshold"
        description: "Your custom metric is {{ $value }}"
```

### Reload Configuration

After editing alert rules:

```powershell
# Reload Prometheus configuration
docker exec prometheus-server kill -HUP 1

# Or restart the service
docker-compose restart prometheus
```

After editing Alertmanager configuration:

```powershell
# Reload Alertmanager configuration
docker exec alertmanager-server kill -HUP 1

# Or restart the service
docker-compose restart alertmanager
```

## Silencing Alerts

### Using Alertmanager UI

1. Go to http://localhost:9093/#/silences
2. Click **New Silence**
3. Set matchers (e.g., `alertname=HighErrorRate`)
4. Set duration
5. Add comment
6. Click **Create**

### Using amtool CLI

```powershell
# Silence specific alert for 2 hours
docker exec alertmanager-server amtool silence add alertname=HighErrorRate -d 2h --comment="Maintenance window"

# List active silences
docker exec alertmanager-server amtool silence query

# Expire a silence
docker exec alertmanager-server amtool silence expire <silence-id>
```

## Alert Routing Logic

```
All Alerts
    │
    ├──> severity=critical ──> critical-receiver (repeat: 1h, wait: 10s)
    │
    ├──> severity=warning ──> warning-receiver (repeat: 4h, wait: 30s)
    │
    └──> default ──> default-receiver (repeat: 4h, wait: 30s)
```

## Grafana Integration

### Add Alertmanager as Data Source

1. Go to Grafana (http://localhost:3000)
2. **Configuration** → **Data Sources** → **Add data source**
3. Select **Alertmanager**
4. Set URL: `http://alertmanager:9093`
5. Click **Save & Test**

### View Alerts in Grafana

1. **Alerting** → **Alert rules**
2. Create dashboards with alert annotations
3. Set up Grafana's own alert rules based on Prometheus queries

## Monitoring Alert System

### Check Prometheus Targets

```
http://localhost:9090/targets
```

Ensure all targets are UP.

### Check Alert Rules

```
http://localhost:9090/rules
```

View all loaded rules and their evaluation status.

### Check Alertmanager Status

```
http://localhost:9093/#/status
```

View Alertmanager configuration and cluster status.

## Troubleshooting

### Alerts Not Firing

1. Check Prometheus can evaluate the rule:
   - Go to http://localhost:9090
   - Enter the alert expression in the query box
   - Verify it returns results

2. Check alert state:
   - Go to http://localhost:9090/alerts
   - Look for PENDING or FIRING status

3. Check evaluation interval:
   - Some alerts have `for: 5m` - they need condition to persist

### Notifications Not Received

1. Check Alertmanager logs:

```powershell
docker logs alertmanager-server
```

2. Verify receiver configuration in `alertmanager.yml`

3. Check webhook/email credentials and network connectivity

4. Look for delivery errors in Alertmanager UI

### Alert Rules Not Loading

1. Check Prometheus logs:

```powershell
docker logs prometheus-server
```

2. Validate YAML syntax:

```powershell
docker exec prometheus-server promtool check rules /etc/prometheus/alert-rules.yml
```

3. Verify file is mounted correctly:

```powershell
docker exec prometheus-server ls -la /etc/prometheus/
```

## Best Practices

1. **Start with Conservative Thresholds**: Adjust based on actual traffic patterns
2. **Use Appropriate Severities**: Critical for issues requiring immediate action
3. **Set Proper Durations**: Avoid alert fatigue from transient issues
4. **Add Meaningful Annotations**: Help responders understand the issue
5. **Test Regularly**: Ensure alerts fire when they should
6. **Use Inhibition Rules**: Prevent duplicate notifications
7. **Document On-Call Procedures**: Include runbooks for each alert
8. **Review and Refine**: Regularly audit alerts and remove noisy ones

## Next Steps

1. **Configure notification channels** (Slack, email, PagerDuty, etc.)
2. **Fine-tune alert thresholds** based on your baseline metrics
3. **Add application-specific alerts** for your business logic
4. **Create runbooks** for each alert type
5. **Set up on-call rotations** with appropriate alert routing
6. **Integrate with incident management** tools (e.g., PagerDuty, Opsgenie)
7. **Create alert dashboards** in Grafana for visualization

## Additional Resources

- [Prometheus Alerting Docs](https://prometheus.io/docs/alerting/latest/overview/)
- [Alertmanager Configuration](https://prometheus.io/docs/alerting/latest/configuration/)
- [Alert Rule Best Practices](https://prometheus.io/docs/practices/alerting/)
- [Grafana Alerting](https://grafana.com/docs/grafana/latest/alerting/)
