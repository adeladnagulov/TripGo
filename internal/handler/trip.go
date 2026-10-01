package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain"
	"github.com/adeladnagulov/TripGo/internal/usecase"
)

type Handler struct {
	tripServise *usecase.TripServise
}

func NewHandler(tripServise *usecase.TripServise) *Handler {
	return &Handler{
		tripServise: tripServise,
	}
}

// CreateTrip Создать поездку
// (POST /api/v1/trips)
func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	var tripData api.TripData
	if err := json.NewDecoder(r.Body).Decode(&tripData); err != nil {
		writeProblem(w, newBadRequest(r.URL.Host, "Invalid request body: "+err.Error()))
		return
	}
	trip, err := h.tripServise.CreateTrip(r.Context(), tripData)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			writeProblem(w, newCreateTripConflict(r.URL.Host, "driver already has active trip"))
			return
		}
		fmt.Println(err) //сделать лог
		writeProblem(w, newInternalError(r.URL.Host, "internal error: "+err.Error()))
		return
	}

	resourceURL := fmt.Sprintf("/api/v1/trips/%s", trip.Id.String())
	w.Header().Set("Location", resourceURL)
	writeJson(w, http.StatusCreated, trip)
}

// GetTrip Получить поездку
// (GET /api/v1/trips/{tripId})
func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	trip, err := h.tripServise.GetTrip(r.Context(), tripId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeProblem(w, newTripNotFound(r.URL.Host, "trip with an id does not exist"))
			return
		}
		fmt.Println(err) //сделать лог
		writeProblem(w, newInternalError(r.URL.Host, "internal error: "+err.Error()))
		return
	}

	writeJson(w, http.StatusOK, trip)
}

// FinishTrip Завершить поездку
// (POST /api/v1/trips/{tripId}/finish)
func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {}

// ListTripPositions Получить маршрут поездки
// (GET /api/v1/trips/{tripId}/positions)
func (h *Handler) ListTripPositions(w http.ResponseWriter, r *http.Request, tripId api.TripId) {}

// CreateTripPosition Сохранить координату поездки
// (POST /api/v1/trips/{tripId}/positions)
func (h *Handler) CreateTripPosition(w http.ResponseWriter, r *http.Request, tripId api.TripId) {}

// Health Liveness
// (GET /health)
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	resp := api.HealthResponse{
		Status: api.Ok,
	}
	writeJson(w, http.StatusOK, resp)
}

// Ready Readiness
// (GET /ready)
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.tripServise.PingTrip(r.Context()); err != nil {
		writeProblem(w, newInternalError(r.URL.Host, "db not respond :"+err.Error()))
		return
	}
	resp := api.HealthResponse{
		Status: api.Ok,
	}
	writeJson(w, http.StatusOK, resp)
}

func newBadRequest(host, detail string) api.BadRequest {
	return api.BadRequest{
		Type:     "about:blank",
		Title:    http.StatusText(http.StatusBadRequest),
		Status:   http.StatusBadRequest,
		Detail:   &detail,
		Code:     "invalid_request",
		Instance: &host,
	}
}

func newCreateTripConflict(host, detail string) api.CreateTripConflict {
	return api.CreateTripConflict{
		Type:     "about:blank",
		Title:    http.StatusText(http.StatusConflict),
		Status:   http.StatusConflict,
		Detail:   &detail,
		Code:     "driver_busy",
		Instance: &host,
	}
}

func newInternalError(host, detail string) api.InternalError {
	return api.InternalError{
		Type:   "about:blank",
		Title:  http.StatusText(http.StatusInternalServerError),
		Status: http.StatusInternalServerError,
		Detail: &detail,
		Code:   "internal_error",
	}
}

func newTripComplited(host, detail string) api.TripCompleted {
	return api.BadRequest{
		Type:     "about:blank",
		Title:    http.StatusText(http.StatusConflict),
		Status:   http.StatusConflict,
		Detail:   &detail,
		Code:     "trip_completed",
		Instance: &host,
	}
}

func newTripNotFound(host, detail string) api.TripNotFound {
	return api.BadRequest{
		Type:     "about:blank",
		Title:    http.StatusText(http.StatusNotFound),
		Status:   http.StatusNotFound,
		Detail:   &detail,
		Code:     "trip_not_found",
		Instance: &host,
	}
}

func writeJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeProblem(w http.ResponseWriter, problem api.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(int(problem.Status))
	json.NewEncoder(w).Encode(problem)
}
