package response

import (
	"encoding/json"
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
)

func Json(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func Problem(w http.ResponseWriter, status int, problem api.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem)
}

func HandleAPIError(w http.ResponseWriter, r *http.Request, err error) {
	status, problem := domainerror.Problem(err, r)
	Problem(w, status, problem)
}
