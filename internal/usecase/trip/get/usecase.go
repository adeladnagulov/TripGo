package get

import (
	"context"
	"fmt"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type GetUseCase struct {
	repo tripGetter
}

func NewGetUseCase(repo tripGetter) *GetUseCase {
	return &GetUseCase{
		repo: repo,
	}
}

func (s *GetUseCase) GetTrip(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get trip id=%v error: %w", id, err)
	}
	return trip, nil
}
