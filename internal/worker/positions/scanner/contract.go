package scanner

import (
	"context"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

type StaleTripsGetter interface {
	GetStaleTrips(ctx context.Context, staleAfter time.Duration) ([]trip.StaleTrip, error)
}

type Ticker interface {
	Chan() <-chan time.Time
	Stop()
}
