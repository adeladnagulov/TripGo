package create

import (
	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

func toDomain(req api.TripData) trip.TripData {
	return trip.TripData{
		DriverId:   req.DriverId,
		UserId:     req.UserId,
		StartPoint: trip.Coordinates{Lat: req.StartPoint.Latitude, Lon: req.StartPoint.Longitude},
		EndPoint:   trip.Coordinates{Lat: req.EndPoint.Latitude, Lon: req.EndPoint.Longitude},
		Price:      req.Price,
	}
}

func toApi(t trip.Trip) api.Trip {
	return api.Trip{
		Id:             t.Id,
		DriverId:       t.DriverId,
		UserId:         t.UserId,
		StartPoint:     api.Coordinates{Latitude: t.StartPoint.Lat, Longitude: t.StartPoint.Lon},
		EndPoint:       api.Coordinates{Latitude: t.EndPoint.Lat, Longitude: t.EndPoint.Lon},
		StartedAt:      t.StartedAt,
		FinishedAt:     t.FinishedAt,
		LastPositionAt: t.LastPositionAt,
		Price:          t.Price,
		Status:         api.TripStatus(t.Status),
	}
}
