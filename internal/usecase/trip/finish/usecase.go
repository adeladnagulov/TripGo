package finish

import (
	"context"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type FinishUseCase struct {
	repo tripFinisher
	tm   txManager
}

func NewCreateUseCase(repo tripFinisher, tm txManager) *FinishUseCase {
	return &FinishUseCase{
		repo: repo,
		tm:   tm,
	}
}

func (s *FinishUseCase) FinishTrip(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	var finishTrip *trip.Trip
	err := s.tm.Do(ctx, func(ctx context.Context) error {
		trip, err := s.repo.Finish(ctx, id, time.Now())
		if err != nil {
			return err
		}
		finishTrip = trip
		return nil
	})
	if err != nil {
		return nil, err
	}
	return finishTrip, nil
}
