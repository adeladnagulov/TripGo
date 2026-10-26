package finish

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type FinishUseCase struct {
	repo   tripFinisher
	tm     txManager
	logger *slog.Logger
}

func NewFinishUseCase(repo tripFinisher, tm txManager, logger *slog.Logger) *FinishUseCase {
	return &FinishUseCase{
		repo:   repo,
		tm:     tm,
		logger: logger,
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
	s.logger.WarnContext(ctx, "trip finished successfully",
		"id", finishTrip.Id,
		"status", finishTrip.Status,
	)
	return finishTrip, nil
}
