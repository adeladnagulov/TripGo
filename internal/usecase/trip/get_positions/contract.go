package getpositions

import (
	"context"

	"github.com/google/uuid"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

type positionsGetter interface {
	GetPositions(ctx context.Context, id uuid.UUID) ([]trip.TripPosition, error)
}
