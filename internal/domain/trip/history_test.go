package trip_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func TestNewTripStatusHistory(t *testing.T) {
	tripID := uuid.New()
	changedAt := time.Date(2026, time.November, 5, 12, 0, 0, 0, time.UTC)
	activeStatus := trip.TripStatusActive

	tests := []struct {
		name       string
		fromStatus *trip.TripStatus
		toStatus   trip.TripStatus
		reason     string
	}{
		{
			name:     "trip created",
			toStatus: trip.TripStatusActive,
			reason:   trip.TripCreatedReason,
		},
		{
			name:       "trip completed",
			fromStatus: &activeStatus,
			toStatus:   trip.TripStatusCompleted,
			reason:     trip.TripFinishedReason,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tripModel := trip.Trip{
				ID:     tripID,
				Status: tt.toStatus,
			}

			history := trip.NewTripStatusHistory(tripModel, tt.fromStatus, tt.reason, changedAt)

			assert.Equal(t, tripID, history.TripID)
			assert.Equal(t, tt.fromStatus, history.FromStatus)
			assert.Equal(t, tt.toStatus, history.ToStatus)
			require.NotNil(t, history.Reason)
			assert.Equal(t, tt.reason, *history.Reason)
			assert.Equal(t, changedAt, history.ChangedAt)
		})
	}
}
