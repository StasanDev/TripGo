package trip_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func TestTripFinish(t *testing.T) {
	finishAt := time.Date(2026, time.November, 5, 12, 0, 0, 0, time.UTC)
	completedAt := finishAt.Add(-time.Hour)

	tests := []struct {
		name           string
		initialStatus  trip.TripStatus
		initialFinish  *time.Time
		wantStatus     trip.TripStatus
		wantFinishedAt *time.Time
		wantErr        error
	}{
		{
			name:           "finish active trip",
			initialStatus:  trip.TripStatusActive,
			wantStatus:     trip.TripStatusCompleted,
			wantFinishedAt: &finishAt,
		},
		{
			name:           "reject completed trip",
			initialStatus:  trip.TripStatusCompleted,
			initialFinish:  &completedAt,
			wantStatus:     trip.TripStatusCompleted,
			wantFinishedAt: &completedAt,
			wantErr:        trip.ErrTripCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tripToFinish := trip.Trip{
				Status:     tt.initialStatus,
				FinishedAt: tt.initialFinish,
			}

			err := tripToFinish.Finish(finishAt)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, tripToFinish.Status)
			if tt.wantFinishedAt == nil {
				assert.Nil(t, tripToFinish.FinishedAt)
				return
			}

			require.NotNil(t, tripToFinish.FinishedAt)
			assert.Equal(t, *tt.wantFinishedAt, *tripToFinish.FinishedAt)
		})
	}
}
