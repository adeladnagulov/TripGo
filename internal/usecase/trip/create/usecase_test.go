package create_test

import (
	"context"
	"testing"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/usecase/trip/create"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestCreateTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(repo *MocktripCreator, tx *MocktxManager)
		input   trip.TripData
		wantErr error
	}{
		{
			name: "success",
			setup: func(repo *MocktripCreator, tx *MocktxManager) {
				tx.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			input: trip.TripData{
				DriverId:   uuid.New(),
				UserId:     uuid.New(),
				Price:      67,
				StartPoint: trip.Coordinates{Lat: 12.12, Lon: 12.12},
				EndPoint:   trip.Coordinates{Lat: 13.13, Lon: 13.13},
			},
			wantErr: nil,
		},
		{
			name: "driver busy",
			setup: func(repo *MocktripCreator, tx *MocktxManager) {
				tx.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(trip.ErrDriverBusy)
			},
			input: trip.TripData{
				DriverId:   uuid.New(),
				UserId:     uuid.New(),
				Price:      67,
				StartPoint: trip.Coordinates{Lat: 12.12, Lon: 12.12},
				EndPoint:   trip.Coordinates{Lat: 13.13, Lon: 13.13},
			},
			wantErr: trip.ErrDriverBusy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMocktripCreator(ctrl)
			mockTx := NewMocktxManager(ctrl)
			tt.setup(mockRepo, mockTx)
			us := create.NewCreateUseCase(mockRepo, mockTx, nil)

			result, err := us.CreateTrip(ctx, tt.input)
			if err != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, result.Status, trip.Active)
			}
		})
	}
}
