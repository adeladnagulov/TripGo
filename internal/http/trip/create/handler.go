package create

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
	"github.com/adeladnagulov/TripGo/internal/http/response"
)

type Handler struct {
	uc tripUseCase
}

func NewHandler(uc tripUseCase) *Handler {
	return &Handler{uc: uc}
}

// CreateTrip Создать поездку
// (POST /api/v1/trips)
func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	var apiTripData api.TripData
	if err := json.NewDecoder(r.Body).Decode(&apiTripData); err != nil {
		status, problem := domainerror.Problem(err, r)
		response.Problem(w, status, problem)
		return
	}
	tripData := toDomain(apiTripData)
	t, err := h.uc.CreateTrip(r.Context(), tripData)
	if err != nil {
		if errors.Is(err, trip.ErrDriverBusy) {
			status, problem := domainerror.Problem(err, r)
			response.Problem(w, status, problem)
			return
		}
		status, problem := domainerror.Problem(err, r)
		response.Problem(w, status, problem)
		return
	}

	resourceURL := fmt.Sprintf("/api/v1/trips/%s", t.Id.String())
	resp := toApi(*t)
	w.Header().Set("Location", resourceURL)
	response.Json(w, http.StatusCreated, resp)
}
