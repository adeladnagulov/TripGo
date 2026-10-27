package finish

import (
	"context"

	"github.com/google/uuid"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

type tripFinisher interface {
	Save(ctx context.Context, t *trip.Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
