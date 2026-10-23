package get

import (
	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/timeutil"
)

func toApi(t trip.Trip) api.Trip {
	return api.Trip{
		Id:             t.Id,
		DriverId:       t.DriverId,
		UserId:         t.UserId,
		StartPoint:     api.Coordinates{Latitude: t.StartPoint.Lat, Longitude: t.StartPoint.Lon},
		EndPoint:       api.Coordinates{Latitude: t.EndPoint.Lat, Longitude: t.EndPoint.Lon},
		StartedAt:      timeutil.FormatTime(t.StartedAt),
		FinishedAt:     timeutil.FormatTimePtr(t.FinishedAt),
		LastPositionAt: timeutil.FormatTimePtr(t.LastPositionAt),
		Price:          t.Price,
		Status:         api.TripStatus(t.Status),
	}
}
