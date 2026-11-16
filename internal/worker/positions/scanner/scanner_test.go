package scanner_test

import (
	"context"
	"log/slog"
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
	"github.com/adeladnagulov/TripGo/internal/worker/positions/scanner"
	workerpool "github.com/adeladnagulov/TripGo/internal/worker/positions/worker_pool"
)

func TestScanner_Deduplication(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := mocks.NewMockStaleTripsGetter(ctrl)
	mockTicker := mocks.NewMockTicker(ctrl)
	mockGateway := mocks.NewMockGateway(ctrl)

	inQueue := currentlyinqueue.NewInQueue()
	logger := slog.Default()

	poolMetrics, _ := metrics.NewPoolMetrics(noop.NewMeterProvider().Meter("test"), func() int { return 0 }, func() int64 { return 0 })
	pool := workerpool.NewWorkerPool(10, inQueue, mockGateway, time.Second, 1, logger, poolMetrics)

	scannerMetrics, err := metrics.NewScannerMetrics(noop.NewMeterProvider().Meter("test"))
	require.NoError(t, err)

	tickerChan := make(chan time.Time)
	mockTicker.EXPECT().Chan().Return(tickerChan).AnyTimes()
	mockTicker.EXPECT().Stop().AnyTimes()

	sc := scanner.NewScanner(30*time.Second, 60*time.Second, mockGetter, pool, inQueue, mockTicker, logger, scannerMetrics)

	ctx, cancel := context.WithCancel(context.Background())
	go sc.Run(ctx)

	trip := domain.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}

	mockGetter.EXPECT().GetStaleTrips(gomock.Any(), 60*time.Second).Return([]domain.StaleTrip{trip, trip}, nil).Times(2)

	tickerChan <- time.Now()
	time.Sleep(50 * time.Millisecond)
	tickerChan <- time.Now()
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, pool.GetQueueSize())

	cancel()
	pool.Stop()
}
