package config

import (
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	HttpAddr                 string        `env:"HTTP_ADDR" envDefault:"8080"`
	LogLevel                 string        `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout          time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	DatabaseUrl              string        `env:"DATABASE_URL,required"`
	DatabaseMaxConns         int           `env:"DATABASE_MAX_CONNS" envDefault:"10"`
	DatabaseMinConns         int           `env:"DATABASE_MIN_CONNS" envDefault:"2"`
	DatabaseMaxConnsLifetime time.Duration `env:"DATABASE_MAX_CONN_LIFETIME" envDefault:"30m"`
	DatabaseConnectTimeout   time.Duration `env:"DATABASE_CONNECT_TIMEOUT" envDefault:"5s"`
	DatabaseQueryTimeout     time.Duration `env:"DATABASE_QUERY_TIMEOUT" envDefault:"3s"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
