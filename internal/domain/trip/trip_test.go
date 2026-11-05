package trip_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func TestNew(t *testing.T) {
	id := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	startedAt := time.Date(2026, time.November, 5, 10, 0, 0, 0, time.UTC)
	input := validTripInput()

	createdTrip, err := trip.NewTrip(input, id, startedAt)

	require.NoError(t, err)
	assert.Equal(t, input, createdTrip.TripInput)
	assert.Equal(t, id, createdTrip.ID)
	assert.Equal(t, trip.TripStatusActive, createdTrip.Status)
	assert.Equal(t, startedAt, createdTrip.StartedAt)
	assert.Nil(t, createdTrip.FinishedAt)
}

func TestNewRejectsInvalidInput(t *testing.T) {
	input := validTripInput()
	input.UserID = uuid.Nil

	createdTrip, err := trip.NewTrip(input, uuid.New(), time.Now())

	var validationErr *trip.ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Equal(t, trip.Trip{}, createdTrip)
}
