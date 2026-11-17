package trip_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func TestNewTripSuccess(t *testing.T) {
	input := validTripInput()
	id := uuid.New()
	startedAt := time.Date(2026, time.November, 5, 10, 0, 0, 0, time.UTC)

	createdTrip, err := trip.NewTrip(input, id, startedAt)

	require.NoError(t, err)
	assert.Equal(t, input, createdTrip.TripInput)
	assert.Equal(t, id, createdTrip.ID)
	assert.Equal(t, startedAt, createdTrip.StartedAt)
	assert.Equal(t, trip.TripStatusActive, createdTrip.Status)
	assert.Nil(t, createdTrip.FinishedAt)
}

func TestNewTripInvalidInput(t *testing.T) {
	input := validTripInput()
	input.UserID = uuid.Nil
	startedAt := time.Date(2026, time.November, 5, 10, 0, 0, 0, time.UTC)

	createdTrip, err := trip.NewTrip(input, uuid.New(), startedAt)

	var validationErr *trip.ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "user_id", validationErr.Field)
	assert.Equal(t, trip.Trip{}, createdTrip)
}
