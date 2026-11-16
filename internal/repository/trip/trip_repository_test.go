//go:build integration

package trip_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adeladnagulov/TripGo/internal/config"
	domain "github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/infra/postgres"
	txManager "github.com/adeladnagulov/TripGo/internal/infra/postgres/transaction"
	"github.com/adeladnagulov/TripGo/internal/repository/trip"
)

func setupTestDb(t *testing.T) (*postgres.DB, func()) {
	t.Helper()
	cfg, err := config.NewConfig()
	require.NoError(t, err, "failed get config")
	db, err := postgres.NewDB(context.Background(), cfg)
	require.NoError(t, err, "failed to connect to DB")
	cleanup := func() {
		db.Close()
	}
	return db, cleanup
}

func TestRepository_Integration(t *testing.T) {
	db, cleanup := setupTestDb(t)
	defer cleanup()

	repo := trip.NewTripRepository(db.Pool, 5*time.Second)
	ctx := context.Background()
	tm := txManager.NewTransactionManager(db.Pool, 5*time.Minute)

	t.Run("create and get trip success", func(t *testing.T) {
		tripId, driverId := uuid.New(), uuid.New()
		newTrip := &domain.Trip{
			Id:         tripId,
			DriverId:   driverId,
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now(),
			Price:      100,
			Status:     domain.Active,
		}
		err := repo.Create(ctx, newTrip)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, tripId)
		require.NoError(t, err)

		assert.Equal(t, newTrip.Id, got.Id)

		t.Cleanup(func() {
			_, _ = db.Pool.Exec(ctx, "DELETE FROM trip_status_history WHERE trip_id = $1", tripId)
			_, _ = db.Pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", tripId)
		})
	})

	t.Run("check atomicity: rollback on history insert failure", func(t *testing.T) {
		tripId, driverId := uuid.New(), uuid.New()

		constraintSQL := fmt.Sprintf(
			"ALTER TABLE trip_status_history ADD CONSTRAINT checker_rollback_test CHECK (trip_id != '%s')",
			tripId.String(),
		)
		_, err := db.Pool.Exec(ctx, constraintSQL)
		require.NoError(t, err, "failed to add a temporary constraint")

		t.Cleanup(func() {
			_, _ = db.Pool.Exec(ctx, `ALTER TABLE trip_status_history DROP CONSTRAINT IF EXISTS checker_rollback_test`)
			_, _ = db.Pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", tripId)
			_, _ = db.Pool.Exec(ctx, "DELETE FROM trip_status_history WHERE trip_id = $1", tripId)
		})

		trip := &domain.Trip{
			Id:         tripId,
			DriverId:   driverId,
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now(),
			Price:      100,
			Status:     domain.Active,
		}

		err = tm.Do(ctx, func(ctx context.Context) error {
			return repo.Create(ctx, trip)
		})
		require.Error(t, err)

		var historyCount int
		err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM trip_status_history WHERE trip_id = $1", tripId).
			Scan(&historyCount)
		require.NoError(t, err)
		require.Equal(t, 0, historyCount)

		var tripCount int
		err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM trips WHERE id = $1", tripId).
			Scan(&tripCount)
		require.NoError(t, err)
		require.Equal(t, 0, tripCount)
	})

	t.Run("successful completion and repetition", func(t *testing.T) {
		tripId, driverId := uuid.New(), uuid.New()
		newTrip := &domain.Trip{
			Id:         tripId,
			DriverId:   driverId,
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now(),
			Price:      100,
			Status:     domain.Active,
		}
		err := repo.Create(ctx, newTrip)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, tripId)
		require.NoError(t, err)

		// success
		err = got.Finish(time.Now())
		require.NoError(t, err)

		err = repo.Save(ctx, got)
		require.NoError(t, err)
		assert.Equal(t, domain.Completed, got.Status)

		newGot, err := repo.GetByID(ctx, tripId)
		require.NoError(t, err)

		// ErrTripCompleted
		err = newGot.Finish(time.Now())
		assert.ErrorIs(t, err, domain.ErrTripCompleted)

		t.Cleanup(func() {
			_, _ = db.Pool.Exec(ctx, "DELETE FROM trip_status_history WHERE trip_id = $1", tripId)
			_, _ = db.Pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", tripId)
		})
	})

	t.Run("obtaining a non existent trip", func(t *testing.T) {
		got, err := repo.GetByID(ctx, uuid.New())
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("GetStaleTrips filters correctly", func(t *testing.T) {
		_, err := db.Pool.Exec(ctx, "DELETE FROM trip_positions")
		require.NoError(t, err, "failed to clean trip_positions")

		_, err = db.Pool.Exec(ctx, "DELETE FROM trips")
		require.NoError(t, err, "failed to clean trips")

		f := time.Now().Add(-1 * time.Hour)
		trip1 := &domain.Trip{
			Id:         uuid.New(),
			DriverId:   uuid.New(),
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now().Add(-2 * time.Hour),
			FinishedAt: &f,
			Price:      100,
			Status:     domain.Completed,
		}
		require.NoError(t, repo.Create(ctx, trip1))

		trip2 := &domain.Trip{
			Id:         uuid.New(),
			DriverId:   uuid.New(),
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now(),
			Price:      100,
			Status:     domain.Active,
		}
		require.NoError(t, repo.Create(ctx, trip2))

		_, err = repo.SavePosition(ctx, domain.TripPosition{TripId: trip2.Id, Coordinates: domain.Coordinates{Lat: 14.12, Lon: 14.12}, RecordedAt: time.Now()})
		require.NoError(t, err)

		// должна попасть
		trip3 := &domain.Trip{
			Id:         uuid.New(),
			DriverId:   uuid.New(),
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now().Add(-2 * time.Hour),
			Price:      100,
			Status:     domain.Active,
		}
		require.NoError(t, repo.Create(ctx, trip3))

		staleTrips, err := repo.GetStaleTrips(ctx, 60*time.Second)
		require.NoError(t, err)

		require.Len(t, staleTrips, 1)
		assert.Equal(t, trip3.Id, staleTrips[0].Id)
		assert.Equal(t, trip3.DriverId, staleTrips[0].DriverId)

		t.Cleanup(func() {
			_, _ = db.Pool.Exec(ctx, "DELETE FROM trip_positions WHERE trip_id IN ($1, $2, $3)", trip1.Id, trip2.Id, trip3.Id)
			_, _ = db.Pool.Exec(ctx, "DELETE FROM trips WHERE id IN ($1, $2, $3)", trip1.Id, trip2.Id, trip3.Id)
		})
	})

	t.Run("SavePosition updates last_position_at atomically and ignores older coords", func(t *testing.T) {
		tripId := uuid.New()
		t.Cleanup(func() { _, _ = db.Pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", tripId) })

		require.NoError(t, repo.Create(ctx, &domain.Trip{
			Id:         tripId,
			DriverId:   uuid.New(),
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now(),
			Price:      100,
			Status:     domain.Active,
		}))

		t1 := time.Now().Add(-10 * time.Minute)
		t2 := time.Now().Add(-5 * time.Minute)
		t0 := time.Now().Add(-20 * time.Minute) // опоздавшая

		err := tm.Do(ctx, func(ctx context.Context) error {
			_, err := repo.SavePosition(ctx, domain.TripPosition{TripId: tripId, Coordinates: domain.Coordinates{Lat: 14.12, Lon: 14.12}, RecordedAt: t1})
			return err
		})
		require.NoError(t, err)

		err = tm.Do(ctx, func(ctx context.Context) error {
			_, err := repo.SavePosition(ctx, domain.TripPosition{TripId: tripId, Coordinates: domain.Coordinates{Lat: 14.12, Lon: 14.12}, RecordedAt: t2})
			return err
		})
		require.NoError(t, err)

		err = tm.Do(ctx, func(ctx context.Context) error {
			_, err := repo.SavePosition(ctx, domain.TripPosition{TripId: tripId, Coordinates: domain.Coordinates{Lat: 14.12, Lon: 14.12}, RecordedAt: t0})
			return err
		})
		require.NoError(t, err)

		var lastPositionAt time.Time
		err = db.Pool.QueryRow(ctx, "SELECT last_position_at FROM trips WHERE id = $1", tripId).Scan(&lastPositionAt)
		require.NoError(t, err)

		assert.Equal(t, t2.Unix(), lastPositionAt.Unix())
	})
}
