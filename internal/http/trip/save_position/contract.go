package createposition

import (
	"context"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type tripUseCase interface {
	SavePosition(ctx context.Context, tripId uuid.UUID, posData trip.PositionData) (*trip.TripPosition, error)
}
