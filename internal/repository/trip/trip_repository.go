package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	txManager "github.com/adeladnagulov/TripGo/internal/infra/postgres/transaction"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewTripRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *Repository {
	return &Repository{
		pool:         pool,
		queryTimeout: queryTimeout,
	}
}

type dbExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func getDbExecutor(ctx context.Context, pool *pgxpool.Pool) dbExecutor {
	if tx, ok := ctx.Value(txManager.TxKey).(pgx.Tx); ok {
		return tx
	}
	return pool
}

func (r *Repository) Ping(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	return r.pool.Ping(queryCtx)
}

func (r *Repository) Create(ctx context.Context, t *trip.Trip) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)

	tripBuilder := sq.Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude", "end_latitude", "end_longitude",
			"price", "status",
			"started_at", "created_at", "updated_at",
		).
		Values(
			t.Id, t.UserId, t.DriverId,
			t.StartPoint.Lat, t.StartPoint.Lon, t.EndPoint.Lat, t.EndPoint.Lon,
			t.Price, t.Status,
			t.StartedAt, time.Now(), time.Now(),
		).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("builde trip query: %w", err)
	}
	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return trip.ErrDriverBusy
		}
		return fmt.Errorf("failed exec trips: %w", err)
	}

	historyBulder := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(t.Id, nil, t.Status, "trip created", t.StartedAt).
		PlaceholderFormat(sq.Dollar)

	query, args, err = historyBulder.ToSql()
	if err != nil {
		return fmt.Errorf("builde trip_status_history query: %w", err)
	}
	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		return fmt.Errorf("failed exec trip_status_history: %w", err)
	}

	return nil
}

func (r *Repository) Save(ctx context.Context, t *trip.Trip) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)

	oldBulder := sq.Select("status").
		From("trips").
		Where(sq.Eq{"id": t.Id}).
		PlaceholderFormat(sq.Dollar)
	query, args, err := oldBulder.ToSql()
	if err != nil {
		return fmt.Errorf("builde old_trip query: %w", err)
	}
	var oldStatus string
	err = dbExc.QueryRow(queryCtx, query, args...).Scan(&oldStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return trip.ErrNotFound
		}
		return fmt.Errorf("failed QueryRow old status: %w", err)
	}

	tripBulder := sq.Update("trips").
		Set("status", t.Status).
		Set("finished_at", t.FinishedAt).
		Set("updated_at", t.FinishedAt).
		Where(sq.Eq{"id": t.Id}).
		PlaceholderFormat(sq.Dollar)
	query, args, err = tripBulder.ToSql()
	if err != nil {
		return fmt.Errorf("builde trip query: %w", err)
	}
	tag, err := dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		return fmt.Errorf("failed exec trips: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return trip.ErrNotFound
	}

	historyBulder := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(t.Id, oldStatus, t.Status, "trip finished", time.Now()). //В будующем, помменять "trip finished" если не только за конца поезди будет
		PlaceholderFormat(sq.Dollar)

	query, args, err = historyBulder.ToSql()
	if err != nil {
		return fmt.Errorf("builde trip_status_history query: %w", err)
	}
	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		return fmt.Errorf("failed exec trip_status_history: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)

	tripBulder := sq.Select(
		"id", "user_id", "driver_id",
		"start_latitude", "start_longitude", "end_latitude", "end_longitude",
		"price", "status",
		"started_at", "finished_at",
	).
		From("trips").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBulder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builde trip query: %w", err)
	}

	t := trip.Trip{}
	err = dbExc.QueryRow(queryCtx, query, args...).Scan(
		&t.Id,
		&t.UserId,
		&t.DriverId,
		&t.StartPoint.Lat,
		&t.StartPoint.Lon,
		&t.EndPoint.Lat,
		&t.EndPoint.Lon,
		&t.Price,
		&t.Status,
		&t.StartedAt,
		&t.FinishedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, trip.ErrNotFound
		}
		return nil, fmt.Errorf("failed QueryRow trip: %w", err)
	}

	return &t, nil
}
