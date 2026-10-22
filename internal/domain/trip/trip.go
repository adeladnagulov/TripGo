package trip

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	Active    Status = "active"
	Completed Status = "completed"
)

type Coordinates struct {
	Lat float64
	Lon float64
}

type Trip struct {
	Id             uuid.UUID
	DriverId       uuid.UUID
	UserId         uuid.UUID
	StartPoint     Coordinates
	EndPoint       Coordinates
	StartedAt      time.Time
	FinishedAt     *time.Time
	LastPositionAt *time.Time
	Price          int64
	Status         Status
}

type TripData struct {
	DriverId   uuid.UUID
	EndPoint   Coordinates
	Price      int64
	StartPoint Coordinates
	UserId     uuid.UUID
}
