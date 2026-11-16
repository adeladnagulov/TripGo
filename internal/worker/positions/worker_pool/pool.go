package workerpool

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	currentlyinqueue "github.com/adeladnagulov/TripGo/internal/worker/positions/currently_in_queue"
	"github.com/adeladnagulov/TripGo/internal/worker/positions/metrics"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var workerTracer = otel.Tracer("worker_pool")

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

	metrics       *metrics.PoolMetrics
	activeWorkers atomic.Int64
	wg            sync.WaitGroup
}

func NewWorkerPool(
	queueSize int,
	inQueue *currentlyinqueue.InQueue,
	pushGateway Gateway,
	pushTimeout time.Duration,
	workers int,
	logger *slog.Logger,
	poolMetrics *metrics.PoolMetrics,
) *WorkerPool {
	return &WorkerPool{
		queue:       make(chan trip.StaleTrip, queueSize),
		inQueue:     inQueue,
		pushGateway: pushGateway,
		pushTimeout: pushTimeout,
		workers:     workers,
		logger:      logger,
		metrics:     poolMetrics,
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
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.queue)
	wp.wg.Wait()
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	wp.logger.DebugContext(ctx, "worker started", slog.Int("worker_id", id))
	defer wp.logger.DebugContext(ctx, "worker stopped", slog.Int("worker_id", id))

	for st := range wp.queue {
		wp.processTrip(ctx, st, id)
	}
}

func (wp *WorkerPool) processTrip(ctx context.Context, st trip.StaleTrip, workerId int) {
	if !wp.inQueue.Contains(st.Id) {
		wp.metrics.RecordRequest(ctx, "cancelled")
		wp.logger.DebugContext(ctx, "task cancelled, skipping",
			slog.String("id", st.Id.String()),
			slog.Int("worker_id", workerId),
		)
		return
	}

	wp.activeWorkers.Add(1)
	defer wp.activeWorkers.Add(-1)

	spanCtx, span := workerTracer.Start(ctx, "worker.process_trip",
		trace.WithLinks(trace.Link{
			SpanContext: st.ScannerSpan,
		}),
	)
	span.SetAttributes(attribute.String("trip.id", st.Id.String()))
	defer span.End()

	pushCtx, cancel := context.WithTimeout(spanCtx, wp.pushTimeout)
	defer cancel()

	start := time.Now()
	err := wp.pushGateway.ReuestPosition(pushCtx, st.Id, st.DriverId)
	duration := time.Since(start)

	wp.inQueue.Remove(st.Id)

	result := "success"
	if err != nil {
		result = "error"
		wp.logger.WarnContext(ctx, "push request failed",
			slog.String("error", err.Error()),
			slog.String("id", st.Id.String()),
			slog.Int("worker_id", workerId),
		)
	} else {
		wp.logger.InfoContext(ctx, "push request sent",
			slog.String("id", st.Id.String()),
			slog.Int("worker_id", workerId),
		)
	}

	wp.metrics.RecordRequest(ctx, result)
	wp.metrics.RecordDuration(ctx, duration, result)
}

func (wp *WorkerPool) GetActiveWorkers() int64 {
	return wp.activeWorkers.Load()
}

func (wp *WorkerPool) GetQueueSize() int {
	return len(wp.queue)
}
