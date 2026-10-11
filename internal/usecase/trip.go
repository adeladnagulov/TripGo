package usecase

import (
	"context"
	"time"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/storage"
	"github.com/google/uuid"
)

type TripServise struct {
	repo storage.TripRepository
	tm   storage.TxManager
}

func NewTripService(repo storage.TripRepository, tm storage.TxManager) *TripServise {
	return &TripServise{
		repo: repo,
		tm:   tm,
	}
}

func (s *TripServise) PingTrip(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *TripServise) CreateTrip(ctx context.Context, data api.TripData) (*api.Trip, error) {
	trip := api.Trip{
		Id:         uuid.New(),
		UserId:     data.UserId,
		DriverId:   data.DriverId,
		StartPoint: data.StartPoint,
		EndPoint:   data.EndPoint,
		Price:      data.Price,
		Status:     api.Active,
		StartedAt:  time.Now(),
	}

	err := s.tm.Do(ctx, func(ctx context.Context) error {
		return s.repo.Create(ctx, &trip)
	})
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (s *TripServise) FinishTrip(ctx context.Context, id uuid.UUID) (*api.Trip, error) {
	var finishTrip *api.Trip
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

func (s *TripServise) GetTrip(ctx context.Context, id uuid.UUID) (*api.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return trip, nil
}
