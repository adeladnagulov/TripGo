package domainerror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/helper/decode"
)

func TestProblem(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/trips/test", nil)
	typeConflict := "about:blank"
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantProblem api.Problem
	}{
		{
			name:        "nil error",
			err:         nil,
			wantStatus:  http.StatusOK,
			wantProblem: api.Problem{},
		},
		{
			name:       "invalid request",
			err:        decode.NewVlidationErr(errors.New("test_error")),
			wantStatus: http.StatusBadRequest,
			wantProblem: api.BadRequest{
				Type:     typeConflict,
				Title:    http.StatusText(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   ptr("invalid request body or parameters"),
				Code:     "invalid_request",
				Instance: ptr("/api/v1/trips/test"),
			},
		},
		{
			name:       "driver busy",
			err:        trip.ErrDriverBusy,
			wantStatus: http.StatusConflict,
			wantProblem: api.CreateTripConflict{
				Type:     typeConflict,
				Title:    http.StatusText(http.StatusConflict),
				Status:   http.StatusConflict,
				Detail:   ptr("driver already has active trip"),
				Code:     "driver_busy",
				Instance: ptr("/api/v1/trips/test"),
			},
		},
		{
			name:       "trip completed",
			err:        trip.ErrTripCompleted,
			wantStatus: http.StatusConflict,
			wantProblem: api.BadRequest{
				Type:     typeConflict,
				Title:    http.StatusText(http.StatusConflict),
				Status:   http.StatusConflict,
				Detail:   ptr("trip is already over"),
				Code:     "trip_completed",
				Instance: ptr("/api/v1/trips/test"),
			},
		},
		{
			name:       "not found",
			err:        trip.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantProblem: api.BadRequest{
				Type:     typeConflict,
				Title:    http.StatusText(http.StatusNotFound),
				Status:   http.StatusNotFound,
				Detail:   ptr("trip with an id does not exist"),
				Code:     "trip_not_found",
				Instance: ptr("/api/v1/trips/test"),
			},
		},
		{
			name:       "internal server error",
			err:        errors.New("db error"),
			wantStatus: http.StatusInternalServerError,
			wantProblem: api.InternalError{
				Type:     typeConflict,
				Title:    http.StatusText(http.StatusInternalServerError),
				Status:   http.StatusInternalServerError,
				Detail:   ptr("internal server error"),
				Code:     "internal_error",
				Instance: ptr("/api/v1/trips/test"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, problem := Problem(tt.err, req)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantProblem, problem)
		})
	}
}

func ptr(s string) *string { return &s }
