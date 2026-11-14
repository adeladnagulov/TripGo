package getpositions

import (
	"context"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type tripUseCase interface {
	GetPositions(ctx context.Context, tripId uuid.UUID) (*trip.TripPositionsResponse, error)
}
