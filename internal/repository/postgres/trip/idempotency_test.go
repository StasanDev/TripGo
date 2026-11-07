//go:build integration

package repository_test

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func (s *RepositorySuite) TestIdempotencyRepositoryReserveNewAndGetExisting() {
	t := s.T()
	ctx := t.Context()

	testTrip := newTestTrip(t)
	err := s.tripRepo.CreateTrip(ctx, testTrip)
	s.Require().NoError(err)

	hash := sha256.Sum256([]byte("first request"))
	now := time.Now().UTC().Truncate(time.Microsecond)
	expected := trip.NewIdempotencyRecord(
		uuid.New(),
		hash[:],
		testTrip.ID,
		now.Add(time.Hour),
	)

	actual, reserved, err := s.idempotencyRepository.ReserveOrGet(ctx, expected)
	s.Require().NoError(err)
	s.True(reserved)
	assertIdempotencyRecordsEqual(t, expected, actual)

	otherHash := sha256.Sum256([]byte("another request"))
	otherRecord := trip.NewIdempotencyRecord(
		expected.Key,
		otherHash[:],
		testTrip.ID,
		expected.ExpiresAt.Add(time.Hour),
	)

	actual, reserved, err = s.idempotencyRepository.ReserveOrGet(ctx, otherRecord)
	s.Require().NoError(err)
	s.False(reserved)
	assertIdempotencyRecordsEqual(t, expected, actual)
}

func (s *RepositorySuite) TestIdempotencyRepositoryReplaceExpired() {
	t := s.T()
	ctx := t.Context()

	firstTrip := newTestTrip(t)
	err := s.tripRepo.CreateTrip(ctx, firstTrip)
	s.Require().NoError(err)

	secondTrip := newTestTrip(t)
	err = s.tripRepo.CreateTrip(ctx, secondTrip)
	s.Require().NoError(err)

	oldHash := sha256.Sum256([]byte("expired request"))
	now := time.Now().UTC().Truncate(time.Microsecond)
	expired := trip.NewIdempotencyRecord(
		uuid.New(),
		oldHash[:],
		firstTrip.ID,
		now.Add(-time.Hour),
	)

	_, reserved, err := s.idempotencyRepository.ReserveOrGet(ctx, expired)
	s.Require().NoError(err)
	s.True(reserved)

	newHash := sha256.Sum256([]byte("new request"))
	expected := trip.NewIdempotencyRecord(
		expired.Key,
		newHash[:],
		secondTrip.ID,
		now.Add(time.Hour),
	)

	actual, reserved, err := s.idempotencyRepository.ReserveOrGet(ctx, expected)
	s.Require().NoError(err)
	s.True(reserved)
	assertIdempotencyRecordsEqual(t, expected, actual)

	actual, reserved, err = s.idempotencyRepository.ReserveOrGet(ctx, expected)
	s.Require().NoError(err)
	s.False(reserved)
	assertIdempotencyRecordsEqual(t, expected, actual)
}

func assertIdempotencyRecordsEqual(
	t *testing.T,
	expected trip.IdempotencyRecord,
	actual trip.IdempotencyRecord,
) {
	t.Helper()

	assert.Equal(t, expected.Key, actual.Key)
	assert.Equal(t, expected.RequestHash, actual.RequestHash)
	assert.Equal(t, expected.TripID, actual.TripID)
	assert.True(t, expected.ExpiresAt.Equal(actual.ExpiresAt))
}
