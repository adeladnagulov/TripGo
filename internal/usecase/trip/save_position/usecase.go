package createposition

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/helper/decode"
	currentlyinqueue "github.com/adeladnagulov/TripGo/internal/worker/positions/currently_in_queue"
)

type SavePositionUseCase struct {
	repo    positionSaver
	tm      txManager
	logger  *slog.Logger
	inQueue *currentlyinqueue.InQueue
	saved   metric.Int64Counter
}

func NewSavePositionUseCase(
	repo positionSaver,
	tm txManager,
	logger *slog.Logger,
	inQueue *currentlyinqueue.InQueue,
	meter metric.Meter,
) *SavePositionUseCase {
	saved, err := meter.Int64Counter(
		"tripgo.positions.saved",
		metric.WithDescription("number of positions saved"),
	)
	if err != nil {
		logger.Error("failed to create saved metric", "error", err)
	}
	return &SavePositionUseCase{
		repo:    repo,
		tm:      tm,
		logger:  logger,
		inQueue: inQueue,
		saved:   saved,
	}
}

func (s *SavePositionUseCase) SavePosition(ctx context.Context, tripId uuid.UUID, posData trip.PositionData) (*trip.TripPosition, error) {
	pos := trip.TripPosition{
		TripId:      tripId,
		Coordinates: posData.Coordinates,
		RecordedAt:  posData.RecordedAt,
	}
	if pos.RecordedAt.After(time.Now().Add(time.Minute)) {
		return nil, decode.NewVlidationErr(trip.ErrRecordedAtInTheFuture)
	}
	err := s.tm.Do(ctx, func(ctx context.Context) error {
		posId, err := s.repo.SavePosition(ctx, pos)
		pos.Id = posId
		return err
	})
	if err != nil {
		s.logger.WarnContext(ctx, "save position error",
			slog.Any("error", err),
			"trip_id", tripId,
			"recorded_at", posData.RecordedAt,
		)
		return nil, fmt.Errorf("save position error: %w", err)
	}

	s.saved.Add(ctx, 1)
	s.inQueue.Remove(tripId)
	s.logger.DebugContext(ctx, "position received, cancelled push request",
		slog.String("trip_id", tripId.String()),
	)

	s.logger.InfoContext(ctx, "save position successfully",
		"id", pos.Id,
		"trip_id", pos.TripId,
		"recorded_at", pos.RecordedAt,
	)
	return &pos, nil
}
