package postgres

import (
	"context"
	"fmt"

	"github.com/adeladnagulov/TripGo/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(ctx context.Context, cfg *config.Config) (*DB, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}
	pgxCfg.MaxConns = int32(cfg.DatabaseMaxConns)
	pgxCfg.MinConns = int32(cfg.DatabaseMinConns)
	pgxCfg.MaxConnLifetime = cfg.DatabaseMaxConnsLifetime
	pgxCfg.ConnConfig.ConnectTimeout = cfg.DatabaseConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), cfg.DatabaseConnectTimeout)
	defer pingCancel()
	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connection from pool: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) PingTrip(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
