//go:build integration

package trip_test

import (
	"context"
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

		_, err := db.Pool.Exec(ctx, `
			ALTER TABLE trip_status_history
			ADD CONSTRAINT checker_rollback_test CHECK (false)
		`)
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
		require.Equal(t, 1, historyCount)

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
}
