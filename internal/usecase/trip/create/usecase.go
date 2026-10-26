package create

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type CreateUseCase struct {
	repo   tripCreator
	tm     txManager
	logger *slog.Logger
}

func NewCreateUseCase(repo tripCreator, tm txManager, logger *slog.Logger) *CreateUseCase {
	return &CreateUseCase{
		repo:   repo,
		tm:     tm,
		logger: logger,
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
		s.logger.WarnContext(ctx, "create trip error",
			"id", t.Id,
			"driver_id", t.DriverId,
		)
		return nil, fmt.Errorf("create trip error: %w", err)
	}
	s.logger.InfoContext(ctx, "trip created successfully",
		"id", t.Id,
		"driver_id", t.DriverId,
		"user_id", t.UserId,
		"price", t.Price,
	)
	return &t, nil
}
