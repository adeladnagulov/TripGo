package workerpool

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	currentlyinqueue "github.com/adeladnagulov/TripGo/internal/worker/positions/currently_in_queue"
	"github.com/adeladnagulov/TripGo/internal/worker/positions/metrics"
)

var workerTracer = otel.Tracer("worker_pool")

type WorkerPool struct {
	queue       chan Task
	inQueue     *currentlyinqueue.InQueue
	pushGateway Gateway
	pushTimeout time.Duration
	workers     int
	logger      *slog.Logger

	metrics       *metrics.PoolMetrics
	activeWorkers atomic.Int64
	wg            sync.WaitGroup
}

type Task struct {
	Trip        trip.StaleTrip
	ScannerSpan trace.SpanContext
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
		queue:       make(chan Task, queueSize),
		inQueue:     inQueue,
		pushGateway: pushGateway,
		pushTimeout: pushTimeout,
		workers:     workers,
		logger:      logger,
		metrics:     poolMetrics,
	}
}

func (wp *WorkerPool) Submit(task Task) bool {
	select {
	case wp.queue <- task:
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

	for task := range wp.queue {
		wp.processTrip(ctx, task, id)
	}
}

func (wp *WorkerPool) processTrip(ctx context.Context, task Task, workerId int) {
	if !wp.inQueue.Contains(task.Trip.Id) {
		wp.metrics.RecordRequest(ctx, "cancelled")
		wp.logger.DebugContext(ctx, "task cancelled, skipping",
			slog.String("id", task.Trip.Id.String()),
			slog.Int("worker_id", workerId),
		)
		return
	}

	wp.activeWorkers.Add(1)
	defer wp.activeWorkers.Add(-1)

	spanCtx, span := workerTracer.Start(ctx, "worker.process_trip",
		trace.WithLinks(trace.Link{
			SpanContext: task.ScannerSpan,
		}),
	)
	span.SetAttributes(attribute.String("trip.id", task.Trip.Id.String()))
	defer span.End()

	pushCtx, cancel := context.WithTimeout(spanCtx, wp.pushTimeout)
	defer cancel()

	start := time.Now()
	err := wp.pushGateway.ReuestPosition(pushCtx, task.Trip.Id, task.Trip.DriverId)
	duration := time.Since(start)

	wp.inQueue.Remove(task.Trip.Id)

	result := "success"
	if err != nil {
		result = "error"
		wp.logger.WarnContext(ctx, "push request failed",
			slog.String("error", err.Error()),
			slog.String("id", task.Trip.Id.String()),
			slog.Int("worker_id", workerId),
		)
	} else {
		wp.logger.InfoContext(ctx, "push request sent",
			slog.String("id", task.Trip.Id.String()),
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
