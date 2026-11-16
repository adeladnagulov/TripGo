package createposition

import (
	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/timeutil"
)

func toDomain(req api.PositionData) trip.PositionData {
	return trip.PositionData{
		Coordinates: trip.Coordinates{Lat: req.Coordinates.Latitude, Lon: req.Coordinates.Longitude},
		RecordedAt:  timeutil.FormatTime(req.RecordedAt),
	}
}

func toApi(pos trip.TripPosition) api.TripPosition {
	return api.TripPosition{
		Id:          pos.Id,
		TripId:      pos.TripId,
		Coordinates: api.Coordinates{Latitude: pos.Coordinates.Lat, Longitude: pos.Coordinates.Lon},
		RecordedAt:  timeutil.FormatTime(pos.RecordedAt),
	}
}
