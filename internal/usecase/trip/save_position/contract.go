package createposition

import (
	"context"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

type positionSaver interface {
	SavePosition(ctx context.Context, pos trip.TripPosition) (int64, error)
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
