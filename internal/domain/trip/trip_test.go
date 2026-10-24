package trip

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTrip_Finish(t *testing.T) {
	at := time.Now()
	tests := []struct {
		name       string // description of this test case
		trip       Trip
		wantStatus Status
		wantErr    error
	}{
		{
			name: "successful completion",
			trip: Trip{
				Status: Active,
			},
			wantStatus: Completed,
			wantErr:    nil,
		},
		{
			name: "has already been completed",
			trip: Trip{
				Status: Completed,
			},
			wantStatus: Completed,
			wantErr:    ErrTripCompleted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trip.Finish(at)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantStatus, tt.trip.Status)
		})
	}
}
