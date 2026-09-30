package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/adeladnagulov/TripGo/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepository interface {
	Create(ctx context.Context, trip *domain.Trip) error
	Finish(ctx context.Context, id uuid.UUID, FinishedAt time.Time) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Trip, error)
	Ping(ctx context.Context) error
}

type tripRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewTripRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *tripRepository {
	return &tripRepository{
		pool:         pool,
		queryTimeout: queryTimeout,
	}
}

type dbExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func getDbExecutor(ctx context.Context, pool *pgxpool.Pool) dbExecutor {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return pool
}

func (r *tripRepository) Ping(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	return r.pool.Ping(queryCtx)
}

func (r *tripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)

	tripBulder := sq.Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude", "end_latitude", "end_longitude",
			"price", "status",
			"started_at", "created_at", "updated_at",
		).
		Values(
			trip.ID, trip.UserID, trip.DriverID,
			trip.StartLatitude, trip.StartLongitude, trip.EndLatitude, trip.EndLongitude,
			trip.Price, trip.Status,
			trip.StartedAt, trip.CreatedAt, trip.UpdatedAt,
		).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBulder.ToSql()
	if err != nil {
		return err
	}
	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return err
	}

	historyBulder := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(trip.ID, nil, trip.Status, "trip created", trip.StartedAt).
		PlaceholderFormat(sq.Dollar)

	query, args, err = historyBulder.ToSql()
	if err != nil {
		return err
	}
	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *tripRepository) Finish(ctx context.Context, id uuid.UUID, finishedAt time.Time) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)

	tripBulder := sq.Update("trips").
		Set("status", domain.TripStatusCompleted).
		Set("finished_at", finishedAt).
		Set("updated_at", finishedAt).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBulder.ToSql()
	if err != nil {
		return err
	}
	result, err := dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("не найдена поездка") //переписать все ошибки
	}

	historyBulder := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(id, domain.TripStatusActive, domain.TripStatusCompleted, "trip finished", finishedAt).
		PlaceholderFormat(sq.Dollar)

	query, args, err = historyBulder.ToSql()
	if err != nil {
		return err
	}
	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *tripRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)

	tripBulder := sq.Select(
		"id", "user_id", "driver_id",
		"start_latitude", "start_longitude", "end_latitude", "end_longitude",
		"price", "status",
		"started_at", "finished_at", "created_at", "updated_at",
	).
		From("trips").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBulder.ToSql()
	if err != nil {
		return nil, err
	}

	trip := domain.Trip{}
	err = dbExc.QueryRow(queryCtx, query, args...).Scan(
		&trip.ID,
		&trip.UserID,
		&trip.DriverID,
		&trip.StartLatitude,
		&trip.StartLongitude,
		&trip.EndLatitude,
		&trip.EndLongitude,
		&trip.Price,
		&trip.Status,
		&trip.StartedAt,
		&trip.FinishedAt,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("не найдена поездка")
		}
		return nil, err
	}

	return &trip, nil
}
