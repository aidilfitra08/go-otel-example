package health

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const healthMetricName = "dependency.health"

var defaultStatuses = map[string]bool{
	"db":    true,
	"redis": true,
}

type Checker struct {
	mu       sync.RWMutex
	statuses map[string]bool
	gauge    metric.Int64ObservableGauge
}

func NewChecker(meter metric.Meter, prefix string) (*Checker, error) {
	gauge, err := meter.Int64ObservableGauge(prefix+healthMetricName, metric.WithDescription("Dependency health status (1=healthy, 0=unhealthy)"))
	if err != nil {
		return nil, err
	}

	checker := &Checker{
		statuses: cloneStatuses(defaultStatuses),
		gauge:    gauge,
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		checker.mu.RLock()
		defer checker.mu.RUnlock()

		for name, healthy := range checker.statuses {
			value := int64(0)
			if healthy {
				value = 1
			}
			observer.ObserveInt64(checker.gauge, value, metric.WithAttributes(
				attribute.String("dependency", name),
			))
		}

		return nil
	}, checker.gauge)
	if err != nil {
		return nil, err
	}

	return checker, nil
}

func (c *Checker) SetStatus(name string, healthy bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statuses[name] = healthy
}

func (c *Checker) Snapshot() (bool, map[string]bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	copy := make(map[string]bool, len(c.statuses))
	overall := true
	for name, healthy := range c.statuses {
		copy[name] = healthy
		if !healthy {
			overall = false
		}
	}

	return overall, copy
}

func cloneStatuses(input map[string]bool) map[string]bool {
	copy := make(map[string]bool, len(input))
	for key, value := range input {
		copy[key] = value
	}

	return copy
}
