package scanner

import (
	"context"
	"log/slog"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	currentlyinqueue "github.com/adeladnagulov/TripGo/internal/worker/positions/currently_in_queue"
	"github.com/adeladnagulov/TripGo/internal/worker/positions/metrics"
	workerpool "github.com/adeladnagulov/TripGo/internal/worker/positions/worker_pool"
)

type Scanner struct {
	checkInterval time.Duration
	staleAfter    time.Duration
	repo          StaleTripsGetter
	workerPool    *workerpool.WorkerPool
	inQueue       *currentlyinqueue.InQueue
	ticker        Ticker
	logger        *slog.Logger
	metrics       *metrics.ScannerMetrics
}

func NewScanner(
	checkInterval time.Duration,
	staleAfter time.Duration,
	repo StaleTripsGetter,
	workerPool *workerpool.WorkerPool,
	inQueue *currentlyinqueue.InQueue,
	ticker Ticker,
	logger *slog.Logger,
	metrics *metrics.ScannerMetrics,

) *Scanner {
	return &Scanner{
		checkInterval: checkInterval,
		staleAfter:    staleAfter,
		repo:          repo,
		workerPool:    workerPool,
		inQueue:       inQueue,
		ticker:        ticker,
		logger:        logger,
		metrics:       metrics,
	}
}

func (s *Scanner) Run(ctx context.Context) {
	s.logger.InfoContext(ctx, "scanner started",
		slog.Duration("check_interval", s.checkInterval),
		slog.Duration("stale_after", s.staleAfter),
	)
	defer func() {
		s.ticker.Stop()
		s.logger.InfoContext(ctx, "scanner stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.ticker.Chan():
			s.tick(ctx)
		}
	}
}

func (s *Scanner) tick(ctx context.Context) {
	staleTrips, err := s.repo.GetStaleTrips(ctx, s.staleAfter)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get stale trips",
			slog.Any("error", err),
		)
		return
	}

	s.logger.DebugContext(ctx, "stale trips fetched",
		slog.Int("count", len(staleTrips)),
	)

	for _, st := range staleTrips {
		s.enqueue(ctx, st)
	}
}

func (s *Scanner) enqueue(ctx context.Context, st trip.StaleTrip) {
	if !s.inQueue.TryAdd(st.Id) {
		s.metrics.RecordCheck(ctx, "duplicate")
		return
	}

	if !s.workerPool.Submit(st) {
		s.inQueue.Remove(st.Id)
		s.metrics.RecordCheck(ctx, "dropped")
		s.logger.WarnContext(ctx, "queue full, task dropped",
			slog.String("id", st.Id.String()),
		)
		return
	}
	s.metrics.RecordCheck(ctx, "queued")
}
