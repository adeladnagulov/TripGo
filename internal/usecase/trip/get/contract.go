package get

import (
	"context"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type tripGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
}
