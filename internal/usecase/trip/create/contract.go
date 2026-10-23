package create

import (
	"context"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

type tripCreator interface {
	Create(ctx context.Context, trip *trip.Trip) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
