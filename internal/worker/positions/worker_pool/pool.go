package workerpool

import (
	"context"
	"log/slog"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	currentlyinqueue "github.com/adeladnagulov/TripGo/internal/worker/positions/currently_in_queue"
	"github.com/google/uuid"
)

type Gateway interface {
	ReuestPosition(ctx context.Context, tripId, driverId uuid.UUID) error
}

type WorkerPool struct {
	queue       chan trip.StaleTrip
	inQueue     *currentlyinqueue.InQueue
	pushGateway Gateway
	pushTimeout time.Duration
	workers     int
	logger      *slog.Logger
}

func NewWorkerPool(
	queueSize int,
	inQueue *currentlyinqueue.InQueue,
	pushGateway Gateway,
	pushTimeout time.Duration,
	workers int,
	logger *slog.Logger,
) *WorkerPool {
	return &WorkerPool{
		queue:       make(chan trip.StaleTrip, queueSize),
		inQueue:     inQueue,
		pushGateway: pushGateway,
		pushTimeout: pushTimeout,
		workers:     workers,
		logger:      logger,
	}
}

func (wp *WorkerPool) Submit(st trip.StaleTrip) bool {
	select {
	case wp.queue <- st:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.queue)
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	wp.logger.DebugContext(ctx, "worker started", slog.Int("worker_id", id))
	defer wp.logger.DebugContext(ctx, "worker stopped", slog.Int("worker_id", id))

	for st := range wp.queue {
		wp.processTrip(ctx, st, id)
	}
}

func (wp *WorkerPool) processTrip(ctx context.Context, st trip.StaleTrip, workerId int) {
	if !wp.inQueue.Contains(st.Id) {
		wp.logger.DebugContext(ctx, "task cancelled, skipping",
			slog.String("id", st.Id.String()),
			slog.Int("worker_id", workerId),
		)
		return
	}

	pushCtx, cancel := context.WithTimeout(ctx, wp.pushTimeout)
	defer cancel()

	err := wp.pushGateway.ReuestPosition(pushCtx, st.Id, st.DriverId)
	wp.inQueue.Remove(st.Id)

	if err != nil {
		wp.logger.WarnContext(ctx, "push request failed",
			slog.String("error", err.Error()),
			slog.String("id", st.Id.String()),
			slog.Int("worker_id", workerId),
		)
		return
	}

	wp.logger.InfoContext(ctx, "push request sent",
		slog.String("id", st.Id.String()),
		slog.Int("worker_id", workerId),
	)
}
