package create

import (
	"context"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

type tripUseCase interface {
	CreateTrip(ctx context.Context, data trip.TripData) (*trip.Trip, error)
}
