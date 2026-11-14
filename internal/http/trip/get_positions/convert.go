package getpositions

import (
	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

func toApi(pos trip.TripPositionsResponse) api.TripPositionsResponse {
	apiTripPosition := make([]api.TripPosition, 0, len(pos.Positions))
	for _, p := range pos.Positions {
		apiPos := api.TripPosition{
			Coordinates: api.Coordinates{Latitude: p.Coordinates.Lat, Longitude: p.Coordinates.Lon},
			Id:          p.Id,
			RecordedAt:  p.RecordedAt,
			TripId:      p.TripId,
		}
		apiTripPosition = append(apiTripPosition, apiPos)
	}
	return api.TripPositionsResponse{
		Positions: apiTripPosition,
	}
}
