-- +goose NO TRANSACTION

-- +goose Up

CREATE INDEX CONCURRENTLY stale_trips_idx ON trips (status, last_position_at)
WHERE status = 'active';

-- +goose Down

DROP INDEX CONCURRENTLY IF EXISTS stale_trips_idx;