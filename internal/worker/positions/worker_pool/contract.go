package workerpool

import (
	"context"

	"github.com/google/uuid"
)

type Gateway interface {
	ReuestPosition(ctx context.Context, tripId, driverId uuid.UUID) error
}
