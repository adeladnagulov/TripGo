package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	txManager "github.com/adeladnagulov/TripGo/internal/infra/postgres/transaction"
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
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
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
	span := trace.SpanFromContext(queryCtx)

	tripBuilder := sq.Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude", "end_latitude", "end_longitude",
			"price", "status",
			"started_at", "finished_at", "created_at", "updated_at",
		).
		Values(
			t.Id, t.UserId, t.DriverId,
			t.StartPoint.Lat, t.StartPoint.Lon, t.EndPoint.Lat, t.EndPoint.Lon,
			t.Price, t.Status,
			t.StartedAt, t.FinishedAt, time.Now(), time.Now(),
		).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("builde trip query: %w", err)
	}

	span.SetAttributes(
		attribute.String("db.operation.name", "INSERT"),
		attribute.String("db.collection.name", "trips"),
	)

	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		span.SetStatus(codes.Error, err.Error())
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

	span.SetAttributes(
		attribute.String("db.operation.name", "INSERT"),
		attribute.String("db.collection.name", "trip_status_history"),
	)

	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed exec trip_status_history: %w", err)
	}

	return nil
}

func (r *Repository) Save(ctx context.Context, t *trip.Trip) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)
	span := trace.SpanFromContext(queryCtx)

	oldBulder := sq.Select("status").
		From("trips").
		Where(sq.Eq{"id": t.Id}).
		PlaceholderFormat(sq.Dollar)
	query, args, err := oldBulder.ToSql()
	if err != nil {
		return fmt.Errorf("builde old_trip query: %w", err)
	}

	span.SetAttributes(
		attribute.String("db.operation.name", "SELECT"),
		attribute.String("db.collection.name", "trips"),
	)

	var oldStatus string
	err = dbExc.QueryRow(queryCtx, query, args...).Scan(&oldStatus)
	if err != nil {
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
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

	span.SetAttributes(
		attribute.String("db.operation.name", "UPDATE"),
		attribute.String("db.collection.name", "trips"),
	)

	tag, err := dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed exec trips: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return trip.ErrTripCompleted
	}

	historyBulder := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(t.Id, oldStatus, t.Status, "trip finished", time.Now()). // В будующем, помменять "trip finished" если не только за конца поезди будет
		PlaceholderFormat(sq.Dollar)

	query, args, err = historyBulder.ToSql()
	if err != nil {
		return fmt.Errorf("builde trip_status_history query: %w", err)
	}

	span.SetAttributes(
		attribute.String("db.operation.name", "INSERT"),
		attribute.String("db.collection.name", "trip_status_history"),
	)

	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed exec trip_status_history: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)
	span := trace.SpanFromContext(queryCtx)

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

	span.SetAttributes(
		attribute.String("db.operation.name", "SELECT"),
		attribute.String("db.collection.name", "trips"),
	)

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
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		span.SetStatus(codes.Error, err.Error())
		if err == pgx.ErrNoRows {
			return nil, trip.ErrNotFound
		}
		return nil, fmt.Errorf("failed QueryRow trip: %w", err)
	}

	return &t, nil
}

func (r *Repository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)
	span := trace.SpanFromContext(queryCtx)

	tripBuilder := sq.Select(
		"id", "user_id", "driver_id",
		"start_latitude", "start_longitude", "end_latitude", "end_longitude",
		"price", "status",
		"started_at", "finished_at",
	).
		From("trips").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		Suffix("FOR UPDATE")

	query, args, err := tripBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builde trip query: %w", err)
	}

	span.SetAttributes(
		attribute.String("db.operation.name", "SELECT"),
		attribute.String("db.collection.name", "trips"),
	)

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
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		span.SetStatus(codes.Error, err.Error())
		if err == pgx.ErrNoRows {
			return nil, trip.ErrNotFound
		}
		return nil, fmt.Errorf("failed QueryRow trip: %w", err)
	}

	return &t, nil
}

func (r *Repository) SavePosition(ctx context.Context, pos trip.TripPosition) (int64, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)
	tripBuilder := sq.Select("status").
		From("trips").
		Where(sq.Eq{"id": pos.TripId}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := tripBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("builde trip query: %w", err)
	}
	var status string
	err = dbExc.QueryRow(queryCtx, query, args...).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, trip.ErrNotFound
		}
		return 0, fmt.Errorf("failed QueryRow trip: %w", err)
	}
	if status != string(trip.Active) {
		return 0, trip.ErrTripCompleted
	}

	positionsBuilder := sq.Insert("trip_positions").
		Columns("trip_id", "latitude", "longitude", "recorded_at").
		Values(pos.TripId, pos.Coordinates.Lat, pos.Coordinates.Lon, pos.RecordedAt).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)
	query, args, err = positionsBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("builde position query: %w", err)
	}
	var posId int64
	err = dbExc.QueryRow(queryCtx, query, args...).Scan(&posId)
	if err != nil {
		return 0, fmt.Errorf("failed insert position: %w", err)
	}

	lastPosBuilder := sq.Update("trips").
		Set("last_position_at", sq.Expr(
			"GREATEST(COALESCE(last_position_at, '1970-01-01 00:00:00+00'::timestamptz), ?)",
			pos.RecordedAt,
		)).
		Where(sq.Eq{"id": pos.TripId}).
		PlaceholderFormat(sq.Dollar)
	query, args, err = lastPosBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("builde last position query: %w", err)
	}
	_, err = dbExc.Exec(queryCtx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed update last position: %w", err)
	}
	return posId, nil
}

func (r *Repository) GetPositions(ctx context.Context, tripId uuid.UUID) ([]trip.TripPosition, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)
	positionsBuilder := sq.Select(
		"id, trip_id, latitude, longitude, recorded_at",
	).
		From("trip_positions").
		Where(sq.Eq{"trip_id": tripId}).
		OrderBy("recorded_at", "id").
		PlaceholderFormat(sq.Dollar)
	query, args, err := positionsBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builde positions query: %w", err)
	}
	rows, err := dbExc.Query(queryCtx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed select positions: %w", err)
	}
	defer rows.Close()
	var positions []trip.TripPosition
	for rows.Next() {
		var pos trip.TripPosition
		err = rows.Scan(&pos.Id, &pos.TripId, &pos.Coordinates.Lat, &pos.Coordinates.Lon, &pos.RecordedAt)
		if err != nil {
			return nil, fmt.Errorf("scan position error: %w", err)
		}
		positions = append(positions, pos)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get positions rows: %w", err)
	}
	return positions, nil
}

func (r *Repository) GetStaleTrips(ctx context.Context, staleAfter time.Duration) ([]trip.StaleTrip, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	dbExc := getDbExecutor(queryCtx, r.pool)
	staleBuilder := sq.Select("id", "driver_id").
		From("trips").
		Where(sq.Eq{"status": trip.Active}).
		Where(sq.Expr(
			"COALESCE(last_position_at, started_at) < ?",
			time.Now().UTC().Add(-staleAfter),
		)).
		OrderBy("COALESCE(last_position_at, started_at)").
		PlaceholderFormat(sq.Dollar)

	query, args, err := staleBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builde stale trips query: %w", err)
	}
	rows, err := dbExc.Query(queryCtx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed execute get stale trips query: %w", err)
	}
	defer rows.Close()
	var staleTrips []trip.StaleTrip
	for rows.Next() {
		var staleTrip trip.StaleTrip
		err = rows.Scan(&staleTrip.Id, &staleTrip.DriverId)
		if err != nil {
			return nil, fmt.Errorf("scan stale trip error: %w", err)
		}
		staleTrips = append(staleTrips, staleTrip)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stale trips rows: %w", err)
	}
	return staleTrips, nil
}
