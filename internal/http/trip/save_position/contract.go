package createposition

import (
	"context"

	"github.com/google/uuid"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

type tripUseCase interface {
	SavePosition(ctx context.Context, tripId uuid.UUID, posData trip.PositionData) (*trip.TripPosition, error)
}
