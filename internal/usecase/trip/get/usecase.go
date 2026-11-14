package get

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

type GetUseCase struct {
	repo   tripGetter
	logger *slog.Logger
}

func NewGetUseCase(repo tripGetter, logger *slog.Logger) *GetUseCase {
	return &GetUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (s *GetUseCase) GetTrip(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WarnContext(ctx, "get trip error",
			slog.String("error", err.Error()),
			"id", id,
		)
		return nil, fmt.Errorf("get trip id=%v error: %w", id, err)
	}
	s.logger.InfoContext(ctx, "trip get successfully",
		"id", trip.Id,
	)
	return trip, nil
}
