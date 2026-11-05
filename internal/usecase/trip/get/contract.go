package get

import (
	"context"

	"github.com/google/uuid"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

type tripGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
}
