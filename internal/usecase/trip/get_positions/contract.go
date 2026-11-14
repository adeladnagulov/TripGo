package getpositions

import (
	"context"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type positionsGetter interface {
	GetPositions(ctx context.Context, id uuid.UUID) ([]trip.TripPosition, error)
}
