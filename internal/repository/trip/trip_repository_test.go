//go:build integration

package trip_test

import (
	"context"
	"testing"
	"time"

	"github.com/adeladnagulov/TripGo/internal/config"
	domain "github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/infra/postgres"
	"github.com/adeladnagulov/TripGo/internal/repository/trip"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			db.Pool.Exec(ctx, "DELETE FROM trip_status_history WHERE trip_id = $1", tripId)
			db.Pool.Exec(ctx, "DELETE FROM trips WHERE trip_id = $1", tripId)
		})
	})

	t.Run("create trip with busy driver AND check atomicity", func(t *testing.T) {
		tripId, driverId := uuid.New(), uuid.New()
		trip1 := &domain.Trip{
			Id:         uuid.New(),
			DriverId:   driverId,
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now(),
			Price:      100,
			Status:     domain.Active,
		}

		err := repo.Create(ctx, trip1)
		require.NoError(t, err)

		var historyCount int
		err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM trip_status_history WHERE trip_id = $1", trip1.Id).
			Scan(&historyCount)
		require.NoError(t, err)
		require.Equal(t, 1, historyCount)

		trip2 := &domain.Trip{
			Id:         uuid.New(),
			DriverId:   driverId,
			UserId:     uuid.New(),
			StartPoint: domain.Coordinates{Lat: 12.12, Lon: 12.12},
			EndPoint:   domain.Coordinates{Lat: 13.13, Lon: 13.13},
			StartedAt:  time.Now(),
			Price:      1000,
			Status:     domain.Active,
		}
		err = repo.Create(ctx, trip2)
		assert.ErrorIs(t, err, domain.ErrDriverBusy)

		var tripCount int
		err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM trips WHERE id = $1", trip2.Id).
			Scan(&tripCount)
		require.NoError(t, err)
		require.Equal(t, 0, tripCount)

		err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM trip_status_history WHERE trip_id = $1", trip2.Id).
			Scan(&historyCount)
		require.NoError(t, err)
		require.Equal(t, 0, historyCount)

		t.Cleanup(func() {
			db.Pool.Exec(ctx, "DELETE FROM trip_status_history WHERE id = $1", tripId)
			db.Pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", tripId)
		})
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
			db.Pool.Exec(ctx, "DELETE FROM trip_status_history WHERE trip_id = $1", tripId)
			db.Pool.Exec(ctx, "DELETE FROM trips WHERE trip_id = $1", tripId)
		})
	})

	t.Run("obtaining a non existent trip", func(t *testing.T) {
		got, err := repo.GetByID(ctx, uuid.New())
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
