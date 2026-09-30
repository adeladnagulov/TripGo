package domain

import (
	"time"

	"github.com/google/uuid"
)

type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TripResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	DriverID       uuid.UUID `json:"driver_id"`
	StartPoint     GeoPoint  `json:"start_point"`
	EndPoint       GeoPoint  `json:"end_point"`
	Price          int64     `json:"price"`
	Status         string    `json:"status"`
	StartedAt      time.Time `json:"started_at"`
	FinishedAt     time.Time `json:"finished_at"`
	LastPositionAt time.Time `json:"last_position_at"`
}

type CreateTripRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	DriverID   uuid.UUID `json:"driver_id"`
	StartPoint GeoPoint  `json:"start_point"`
	EndPoint   GeoPoint  `json:"end_point"`
	Price      int64     `json:"price"`
}
