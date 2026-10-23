package create

import (
	"context"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type CreateUseCase struct {
	repo tripCreator
	tm   txManager
}

func NewCreateUseCase(repo tripCreator, tm txManager) *CreateUseCase {
	return &CreateUseCase{
		repo: repo,
		tm:   tm,
	}
}

func (s *CreateUseCase) CreateTrip(ctx context.Context, data trip.TripData) (*trip.Trip, error) {
	t := trip.Trip{
		Id:         uuid.New(),
		UserId:     data.UserId,
		DriverId:   data.DriverId,
		StartPoint: data.StartPoint,
		EndPoint:   data.EndPoint,
		Price:      data.Price,
		Status:     trip.Active,
		StartedAt:  time.Now(),
	}

	err := s.tm.Do(ctx, func(ctx context.Context) error {
		return s.repo.Create(ctx, &t)
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}
