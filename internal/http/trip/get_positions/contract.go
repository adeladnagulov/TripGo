package getpositions

import (
	"context"

	"github.com/google/uuid"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

type tripUseCase interface {
	GetPositions(ctx context.Context, tripId uuid.UUID) (*trip.TripPositionsResponse, error)
}
