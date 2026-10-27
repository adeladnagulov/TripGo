package get_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/usecase/trip/get"
)

func TestGetTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(repo *MocktripGetter)
		wantErr error
	}{
		{
			name: "success",
			setup: func(repo *MocktripGetter) {
				repo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(&trip.Trip{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "trip not found",
			setup: func(repo *MocktripGetter) {
				repo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(nil, trip.ErrNotFound)
			},
			wantErr: trip.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMocktripGetter(ctrl)
			tt.setup(mockRepo)
			us := get.NewGetUseCase(mockRepo, slog.Default())

			_, err := us.GetTrip(ctx, uuid.New())
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
