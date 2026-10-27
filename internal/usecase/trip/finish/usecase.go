package finish

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

type FinishUseCase struct {
	repo           tripFinisher
	tm             txManager
	logger         *slog.Logger
	tripsComplited metric.Int64Counter
}

func NewFinishUseCase(repo tripFinisher, tm txManager, logger *slog.Logger, meter metric.Meter) *FinishUseCase {
	tripsComplited, err := meter.Int64Counter(
		"tripgo.trips.complete",
		metric.WithDescription("completed trips"),
	)
	if err != nil {
		logger.Error("failed to create trips_complited metric", "error", err)
	}
	return &FinishUseCase{
		repo:           repo,
		tm:             tm,
		logger:         logger,
		tripsComplited: tripsComplited,
	}
}

func (s *FinishUseCase) FinishTrip(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	finishTrip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get finish trip id=%v error: %w", id, err)
	}

	err = finishTrip.Finish(time.Now())
	if err != nil {
		s.logger.WarnContext(ctx, "trip already complited",
			"id", finishTrip.Id,
		)
		return nil, err
	}

	err = s.tm.Do(ctx, func(ctx context.Context) error {
		err := s.repo.Save(ctx, finishTrip)
		if err != nil {
			return fmt.Errorf("save finish trip id=%v error: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.tripsComplited.Add(ctx, 1)
	s.logger.WarnContext(ctx, "trip finished successfully",
		"id", finishTrip.Id,
		"status", finishTrip.Status,
	)
	return finishTrip, nil
}
