package get_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/trip/get"
)

func TestGetTrip(t *testing.T) {
	tripId := uuid.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		setup      func(uc *MocktripUseCase)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					GetTrip(gomock.Any(), gomock.Any()).
					Return(&trip.Trip{Id: tripId}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "trip not found",
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					GetTrip(gomock.Any(), gomock.Any()).
					Return(nil, trip.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "trip_not_found",
		},
		{
			name: "internal error",
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					GetTrip(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUc := NewMocktripUseCase(ctrl)
			tt.setup(mockUc)
			handler := get.NewHandler(mockUc)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/api/v1/trips/676767", bytes.NewBufferString(""))
			r.Header.Set("Content-Type", "application/json")

			handler.GetTrip(w, r, tripId)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCode != "" {
				var problem api.Problem
				err := json.NewDecoder(w.Body).Decode(&problem)
				require.NoError(t, err)
				assert.Equal(t, tt.wantCode, problem.Code)
			}
		})
	}
}
