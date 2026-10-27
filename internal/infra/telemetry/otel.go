package telemetry

import (
	"context"
	"log/slog"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/adeladnagulov/TripGo/internal/config"
)

type OTelProviders struct {
	Tracer *trace.TracerProvider
	Meter  *metric.MeterProvider
}

func NewOTel(ctx context.Context, cfg *config.Config) (*OTelProviders, error) {
	p := &OTelProviders{}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.OtelServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	endpointHost := cfg.OtelExporterEndpoint
	if u, err := url.Parse(cfg.OtelExporterEndpoint); err == nil && u.Host != "" {
		endpointHost = u.Host
	}

	traceExp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpointHost),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		slog.Warn("trace exporter init failed, traces disabled", "error", err)
	} else {
		p.Tracer = trace.NewTracerProvider(
			trace.WithResource(res),
			trace.WithBatcher(traceExp),
			trace.WithSampler(trace.ParentBased(trace.AlwaysSample())),
		)
		otel.SetTracerProvider(p.Tracer)
	}

	metricExp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpointHost),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		slog.Warn("metric exporter init failed", "error", err)
	} else {
		p.Meter = metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(metric.NewPeriodicReader(metricExp)),
		)
		otel.SetMeterProvider(p.Meter)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return p, nil
}

func (tp *OTelProviders) Shutdown(ctx context.Context) {
	if tp.Tracer != nil {
		_ = tp.Tracer.Shutdown(ctx)
	}
	if tp.Meter != nil {
		_ = tp.Meter.Shutdown(ctx)
	}
}
