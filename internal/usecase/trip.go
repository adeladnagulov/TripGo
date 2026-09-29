package usecase

import (
	"context"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain"
	"github.com/adeladnagulov/TripGo/internal/storage"
	"github.com/google/uuid"
)

type TripServise struct {
	repo storage.TripRepository
	tm   storage.TxManager
}

func NewTripServise(repo storage.TripRepository, tm storage.TxManager) *TripServise {
	return &TripServise{
		repo: repo,
		tm:   tm,
	}
}

func (s *TripServise) CreateTrip(ctx context.Context, req domain.CreateTripRequest) (*domain.Trip, error) {
	trip := domain.Trip{
		ID:             uuid.New(),
		UserID:         req.UserID,
		DriverID:       req.DriverID,
		StartLatitude:  req.StartLatitude,
		StartLongitude: req.StartLongitude,
		EndLatitude:    req.EndLatitude,
		EndLongitude:   req.EndLongitude,
		Price:          req.Price,
		Status:         "active",
		StartedAt:      time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.tm.Do(ctx, func(ctx context.Context) error {
		return s.repo.Create(ctx, &trip)
	})
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (s *TripServise) FinishTrip(ctx context.Context, id uuid.UUID) error {
	err := s.tm.Do(ctx, func(ctx context.Context) error {
		return s.repo.Finish(ctx, id, time.Now())
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *TripServise) GetTrip(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return trip, nil
}
