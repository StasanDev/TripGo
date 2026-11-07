//go:build integration

package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func (s *RepositorySuite) TestTripRepositoryGetNotFound() {
	t := s.T()
	got, err := s.tripRepo.GetTrip(t.Context(), uuid.New())
	s.Require().ErrorIs(err, trip.ErrTripNotFound)
	s.Empty(got)
}

func (s *RepositorySuite) TestTripRepositoryCreateAndGet() {
	t := s.T()
	testTrip := newTestTrip(t)
	err := s.tripRepo.CreateTrip(t.Context(), testTrip)
	s.Require().NoError(err)
	got, err := s.tripRepo.GetTrip(t.Context(), testTrip.ID)
	s.Require().NoError(err)
	assertTripsEqual(t, testTrip, got)
}

func (s *RepositorySuite) TestNewTripOnActiveDriver() {
	t := s.T()
	trip1 := newTestTrip(t)
	trip2 := newTestTrip(t)
	trip2.DriverID = trip1.DriverID
	err := s.tripRepo.CreateTrip(t.Context(), trip1)
	s.Require().NoError(err)
	err = s.tripRepo.CreateTrip(t.Context(), trip2)
	s.Require().ErrorIs(err, trip.ErrDriverBusy)
}

func (s *RepositorySuite) TestTripRepositoryFinish() {
	t := s.T()
	testTrip := newTestTrip(t)
	err := s.tripRepo.CreateTrip(t.Context(), testTrip)
	s.Require().NoError(err)

	at := time.Date(2026, time.November, 7, 10, 0, 0, 0, time.UTC)
	testTrip.FinishedAt = &at
	testTrip.Status = trip.TripStatusCompleted
	got, err := s.tripRepo.FinishTrip(t.Context(), testTrip.ID, at)
	s.Require().NoError(err)
	assertTripsEqual(t, testTrip, got)
}

func (s *RepositorySuite) TestTripRepositoryFinishAlreadyCompleted() {
	t := s.T()
	testTrip := newTestTrip(t)
	err := s.tripRepo.CreateTrip(t.Context(), testTrip)
	s.Require().NoError(err)

	at := time.Date(2026, time.November, 7, 10, 0, 0, 0, time.UTC)
	_, err = s.tripRepo.FinishTrip(t.Context(), testTrip.ID, at)
	s.Require().NoError(err)
	got, err := s.tripRepo.FinishTrip(t.Context(), testTrip.ID, at)
	s.Require().ErrorIs(err, trip.ErrTripCompleted)
	s.Empty(got)
}

func newTestTrip(t *testing.T) trip.Trip {
	t.Helper()

	testTrip, err := trip.NewTrip(
		trip.TripInput{
			UserID:   uuid.New(),
			DriverID: uuid.New(),
			StartPoint: trip.Coordinates{
				Latitude:  55.751244,
				Longitude: 37.618423,
			},
			EndPoint: trip.Coordinates{
				Latitude:  55.755814,
				Longitude: 37.617635,
			},
			Price: 500,
		},
		uuid.New(),
		time.Date(2026, time.November, 6, 10, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)

	return testTrip
}

func assertTripsEqual(t *testing.T, expected, actual trip.Trip) {
	t.Helper()

	assert.Equal(t, expected.TripInput, actual.TripInput)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Status, actual.Status)
	assert.True(t, expected.StartedAt.Equal(actual.StartedAt))

	if expected.FinishedAt == nil {
		assert.Nil(t, actual.FinishedAt)
		return
	}

	require.NotNil(t, actual.FinishedAt)
	assert.True(t, expected.FinishedAt.Equal(*actual.FinishedAt))
}
