package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	HttpAddr                 string        `env:"HTTP_ADDR" envDefault:"8080"`
	HttpReadTimeout          time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	HttpReadHeaderTimeout    time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	HttpWriteTimeout         time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	HttpIdleTimeout          time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	LogLevel                 string        `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout          time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	DatabaseUrl              string        `env:"DATABASE_URL,required"`
	DatabaseMaxConns         int           `env:"DATABASE_MAX_CONNS" envDefault:"10"`
	DatabaseMinConns         int           `env:"DATABASE_MIN_CONNS" envDefault:"2"`
	DatabaseMaxConnsLifetime time.Duration `env:"DATABASE_MAX_CONN_LIFETIME" envDefault:"30m"`
	DatabaseConnectTimeout   time.Duration `env:"DATABASE_CONNECT_TIMEOUT" envDefault:"5s"`
	DatabaseQueryTimeout     time.Duration `env:"DATABASE_QUERY_TIMEOUT" envDefault:"3s"`
	OtelServiceName          string        `env:"OTEL_SERVICE_NAME" envDefault:"trip-service"`
	OtelExporterEndpoint     string        `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"http://localhost:22318"`
	OtelExporterProtocol     string        `env:"OTEL_EXPORTER_OTLP_PROTOCOL" envDefault:"http/protobuf"`
	OtelResourceAttrs        string        `env:"OTEL_RESOURCE_ATTRIBUTES" envDefault:"service.namespace=tripgo,deployment.environment=local"`
	OtelTracesSampler        string        `env:"OTEL_TRACES_SAMPLER" envDefault:"parentbased_always_on"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func ParseLodLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
