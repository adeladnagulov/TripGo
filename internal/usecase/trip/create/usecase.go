package create

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"
)

type CreateUseCase struct {
	repo         tripCreator
	tm           txManager
	logger       *slog.Logger
	meter        metric.Meter
	tripsCreated metric.Int64Counter
}

func NewCreateUseCase(repo tripCreator, tm txManager, logger *slog.Logger, meter metric.Meter) *CreateUseCase {
	tripsCreated, err := meter.Int64Counter(
		"tripgo.trips.created",
		metric.WithDescription("created trips"),
	)
	if err != nil {
		logger.Error("failed to create trips_created metric", "error", err)
	}
	return &CreateUseCase{
		repo:         repo,
		tm:           tm,
		logger:       logger,
		meter:        meter,
		tripsCreated: tripsCreated,
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
	s.tripsCreated.Add(ctx, 1)
	s.logger.InfoContext(ctx, "trip created successfully",
		"id", t.Id,
		"driver_id", t.DriverId,
		"user_id", t.UserId,
		"price", t.Price,
	)
	return &t, nil
}
