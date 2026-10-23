package finish

import (
	"context"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type tripFinisher interface {
	Finish(ctx context.Context, id uuid.UUID, FinishedAt time.Time) (*trip.Trip, error)
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
