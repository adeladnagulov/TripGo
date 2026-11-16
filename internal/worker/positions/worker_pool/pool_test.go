package workerpool_test

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	gomock "go.uber.org/mock/gomock"

	domain "github.com/adeladnagulov/TripGo/internal/domain/trip"
	currentlyinqueue "github.com/adeladnagulov/TripGo/internal/worker/positions/currently_in_queue"
	"github.com/adeladnagulov/TripGo/internal/worker/positions/metrics"
	"github.com/adeladnagulov/TripGo/internal/worker/positions/mocks"
	workerpool "github.com/adeladnagulov/TripGo/internal/worker/positions/worker_pool"
)

func setupPoolTest(t *testing.T) (*gomock.Controller, *currentlyinqueue.InQueue, *mocks.MockGateway, *metrics.PoolMetrics, *slog.Logger) {
	t.Helper()
	ctrl := gomock.NewController(t)
	inQueue := currentlyinqueue.NewInQueue()

	mockGateway := mocks.NewMockGateway(ctrl)
	logger := slog.Default()

	poolMetrics, err := metrics.NewPoolMetrics(
		noop.NewMeterProvider().Meter("test"),
		func() int { return 0 },
		func() int64 { return 0 },
	)
	require.NoError(t, err)

	return ctrl, inQueue, mockGateway, poolMetrics, logger
}

func TestWorkerPool_Scenarios(t *testing.T) {
	ctrl, inQueue, mockGateway, poolMetrics, logger := setupPoolTest(t)
	defer ctrl.Finish()

	t.Run("Queue overflow gives dropped, no blocking", func(t *testing.T) {
		pool := workerpool.NewWorkerPool(1, inQueue, mockGateway, time.Second, 0, logger, poolMetrics)

		task1 := domain.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}
		task2 := domain.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}

		assert.True(t, pool.Submit(task1))
		assert.False(t, pool.Submit(task2))
	})

	t.Run("Cancellation by coordinate arrival", func(t *testing.T) {
		tripID := uuid.New()

		inQueue.TryAdd(tripID)
		inQueue.Remove(tripID)

		mockGateway.EXPECT().ReuestPosition(gomock.Any(), tripID, gomock.Any()).Times(0)

		pool := workerpool.NewWorkerPool(10, inQueue, mockGateway, time.Second, 1, logger, poolMetrics)
		ctx, cancel := context.WithCancel(context.Background())
		pool.Start(ctx)

		pool.Submit(domain.StaleTrip{Id: tripID, DriverId: uuid.New()})

		assert.Eventually(t, func() bool {
			return true
		}, 100*time.Millisecond, 10*time.Millisecond)

		cancel()
		pool.Stop()
	})

	t.Run("Push Service error doesn't crash worker", func(t *testing.T) {
		var callCount atomic.Int32

		mockGateway.EXPECT().ReuestPosition(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, tripId, driverId uuid.UUID) error {
				count := callCount.Add(1)
				if count == 1 {
					return errors.New("push service timeout")
				}
				return nil
			}).Times(2)

		pool := workerpool.NewWorkerPool(10, inQueue, mockGateway, time.Second, 1, logger, poolMetrics)
		ctx, cancel := context.WithCancel(context.Background())
		pool.Start(ctx)

		trip1 := domain.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}
		trip2 := domain.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}

		inQueue.TryAdd(trip1.Id)
		inQueue.TryAdd(trip2.Id)

		pool.Submit(trip1)
		pool.Submit(trip2)

		assert.Eventually(t, func() bool {
			return callCount.Load() == 2
		}, 2*time.Second, 10*time.Millisecond, "Worker crashed or hung, did not process the second task")

		cancel()
		pool.Stop()
	})

	t.Run("Context cancellation stops workers, no goroutines left", func(t *testing.T) {
		mockGateway.EXPECT().ReuestPosition(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

		pool := workerpool.NewWorkerPool(10, inQueue, mockGateway, time.Second, 2, logger, poolMetrics)
		ctx, cancel := context.WithCancel(context.Background())
		pool.Start(ctx)

		cancel()
		pool.Stop()
	})
}
