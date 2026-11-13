package server

import (
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/http/health"
	"github.com/adeladnagulov/TripGo/internal/http/trip/create"
	"github.com/adeladnagulov/TripGo/internal/http/trip/finish"
	"github.com/adeladnagulov/TripGo/internal/http/trip/get"
	createposition "github.com/adeladnagulov/TripGo/internal/http/trip/save_position"
)

type Server struct {
	getHandler       *get.Handler
	createHandlet    *create.Handler
	finishHandler    *finish.Handler
	healthHandler    *health.Handler
	createPosHandler *createposition.Handler
}

func NewServerInterface(
	getHandler *get.Handler,
	createHandlet *create.Handler,
	finishHandler *finish.Handler,
	healthHandler *health.Handler,
	createPosHandler *createposition.Handler,
) *Server {
	return &Server{
		getHandler:       getHandler,
		createHandlet:    createHandlet,
		finishHandler:    finishHandler,
		healthHandler:    healthHandler,
		createPosHandler: createPosHandler,
	}
}

func (s *Server) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	s.createHandlet.CreateTrip(w, r, params)
}

func (s *Server) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	s.getHandler.GetTrip(w, r, tripId)
}

func (s *Server) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	s.finishHandler.FinishTrip(w, r, tripId)
}

func (s *Server) ListTripPositions(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) CreateTripPosition(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	s.createPosHandler.CreateTripPosition(w, r, tripId)
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	s.healthHandler.Health(w, r)
}

func (s *Server) Ready(w http.ResponseWriter, r *http.Request) {
	s.healthHandler.Ready(w, r)
}
