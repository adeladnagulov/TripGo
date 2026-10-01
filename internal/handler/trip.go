package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/adeladnagulov/TripGo/internal/domain"
	"github.com/adeladnagulov/TripGo/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	tripServise *usecase.TripServise
}

func newHandler(tripServise *usecase.TripServise) *Handler {
	return &Handler{
		tripServise: tripServise,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.tripServise.PingTrip(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status": "error", "reason": "database_unreachable"}`))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ready"}`))
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var createTripRequest domain.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&createTripRequest); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}
	trip, err := h.tripServise.CreateTrip(r.Context(), createTripRequest)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			writeError(w, r, http.StatusConflict, "driver_busy", "driver already has active trip")
			return
		}
		fmt.Println(err) //сделать лог
		writeError(w, r, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	tripResp := domain.TripResponse{
		ID:       trip.ID,
		UserID:   trip.UserID,
		DriverID: trip.DriverID,
		StartPoint: domain.GeoPoint{
			Latitude:  trip.StartLatitude,
			Longitude: trip.StartLongitude,
		},
		EndPoint: domain.GeoPoint{
			Latitude:  trip.EndLatitude,
			Longitude: trip.EndLongitude,
		},
		Price:     trip.Price,
		Status:    trip.Status,
		StartedAt: trip.CreatedAt,
	}
	resourceURL := fmt.Sprintf("/api/v1/trips/%s", tripResp.ID.String())
	w.Header().Set("Location", resourceURL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tripResp)
}

// 605b1f88-0102-48dc-96bf-fb5490ed9ea2
func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "trip_not_found", "trip with an id does not exist")
		return
	}

	trip, err := h.tripServise.GetTrip(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "trip_not_found", "trip with an id does not exist")
			return
		}
		fmt.Println(err) //сделать лог
		writeError(w, r, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	tripResp := domain.TripResponse{
		ID:       trip.ID,
		UserID:   trip.UserID,
		DriverID: trip.DriverID,
		StartPoint: domain.GeoPoint{
			Latitude:  trip.StartLatitude,
			Longitude: trip.StartLongitude,
		},
		EndPoint: domain.GeoPoint{
			Latitude:  trip.EndLatitude,
			Longitude: trip.EndLongitude,
		},
		Price:     trip.Price,
		Status:    trip.Status,
		StartedAt: trip.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tripResp)
}

func writeError(w http.ResponseWriter, r *http.Request, starus int, code string, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(starus)
	problem := domain.Problem{
		Type:     "about:blank",
		Title:    http.StatusText(starus),
		Status:   starus,
		Detail:   detail,
		Code:     code,
		Instance: r.URL.Path,
	}
	json.NewEncoder(w).Encode(problem)
}
