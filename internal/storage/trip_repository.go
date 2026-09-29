package storage

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/adeladnagulov/TripGo/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepository interface {
	Create(ctx context.Context, trip *domain.Trip) error
	Finish(ctx context.Context, id uuid.UUID, FinishedAt time.Duration) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Trip, error)
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

func (r *tripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	tripBulder := sq.Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude", "end_latitude", "end_longitude",
			"price", "status",
			"started_at", "created_at", "updated_at",
		).
		Values(
			trip.ID, trip.UserID, trip.DriverID,
			trip.StartLatitude, trip.EndLongitude, trip.EndLatitude, trip.EndLongitude,
			trip.Price, trip.Status,
			trip.StartedAt, trip.CreatedAt, trip.UpdatedAt,
		).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBulder.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(queryCtx, query, args...)
	if err != nil {
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
	_, err = r.pool.Exec(queryCtx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *tripRepository) Finish(ctx context.Context, id uuid.UUID, finishedAt time.Duration) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	tripBulder := sq.Update("trips").
		Set("status", "completed").
		Set("finished_at", finishedAt).
		Set("updated_at", finishedAt).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBulder.ToSql()
	if err != nil {
		return err
	}
	result, err := r.pool.Exec(queryCtx, query, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("не найдена поездка") //переписать все ошибки
	}

	historyBulder := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(id, "active", "completed", "trip finished", finishedAt).
		PlaceholderFormat(sq.Dollar)

	query, args, err = historyBulder.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(queryCtx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *tripRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

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
	err = r.pool.QueryRow(queryCtx, query, args...).Scan(
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
