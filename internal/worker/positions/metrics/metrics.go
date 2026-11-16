package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ScannerMetrics struct {
	staleChecks metric.Int64Counter
}

func NewScannerMetrics(meter metric.Meter) (*ScannerMetrics, error) {
	counter, err := meter.Int64Counter(
		"tripgo.stale_checks",
		metric.WithDescription("что сканер сделал с поездкой на этом тике"),
	)
	if err != nil {
		return nil, err
	}
	return &ScannerMetrics{staleChecks: counter}, nil
}

func (m *ScannerMetrics) RecordCheck(ctx context.Context, result string) {
	m.staleChecks.Add(ctx, 1, metric.WithAttributes(attribute.String("result", result)))
}

type PoolMetrics struct {
	requests metric.Int64Counter
	duration metric.Float64Histogram
}

func NewPoolMetrics(
	meter metric.Meter,
	getQueueSize func() int,
	getActiveWorkers func() int64,
) (*PoolMetrics, error) {
	requests, err := meter.Int64Counter(
		"tripgo.position_requests",
		metric.WithDescription("чем закончилась попытка запросить координату"),
	)
	if err != nil {
		return nil, err
	}

	duration, err := meter.Float64Histogram(
		"tripgo.position_requests.duration",
		metric.WithDescription("сколько занял вызов Push Service"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	_, err = meter.Int64ObservableGauge(
		"tripgo.position_requests.queue.size",
		metric.WithDescription("сколько задач сейчас в очереди"),
		metric.WithInt64Callback(func(ctx context.Context, obsrv metric.Int64Observer) error {
			obsrv.Observe(int64(getQueueSize()))
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	_, err = meter.Int64ObservableGauge(
		"tripgo.position_requests.workers.active",
		metric.WithDescription("сколько воркеров сейчас заняты"),
		metric.WithInt64Callback(func(ctx context.Context, obsrv metric.Int64Observer) error {
			obsrv.Observe(getActiveWorkers())
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return &PoolMetrics{
		requests: requests,
		duration: duration,
	}, nil
}

func (m *PoolMetrics) RecordRequest(ctx context.Context, result string) {
	m.requests.Add(ctx, 1, metric.WithAttributes(attribute.String("result", result)))
}

func (m *PoolMetrics) RecordDuration(ctx context.Context, duration time.Duration, result string) {
	m.duration.Record(ctx, duration.Seconds(), metric.WithAttributes(attribute.String("result", result)))
}
