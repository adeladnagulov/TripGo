package finish

import (
	"context"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

type tripUseCase interface {
	FinishTrip(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
}
