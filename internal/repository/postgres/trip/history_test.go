//go:build integration

package repository_test

import (
	"context"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func (s *RepositorySuite) TestAddInHistoryCreate() {
	t := s.T()
	testTrip := newTestTrip(t)
	err := s.tripRepo.CreateTrip(t.Context(), testTrip)
	s.Require().NoError(err)

	expected := trip.NewTripStatusHistory(
		testTrip,
		nil,
		trip.TripCreatedReason,
		testTrip.StartedAt,
	)

	err = s.tripStatusHistoryRepository.AddInTripHistory(t.Context(), expected)
	s.Require().NoError(err)

	actual := getTripHistory(t, t.Context(), s.infra.Pool, testTrip.ID)

	s.Equal(expected.TripID, actual.TripID)
	s.Nil(actual.FromStatus)
	s.Equal(expected.ToStatus, actual.ToStatus)
	s.Equal(expected.Reason, actual.Reason)
	s.True(expected.ChangedAt.Equal(actual.ChangedAt))
}

func getTripHistory(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	tripID uuid.UUID,
) trip.TripStatusHistory {
	t.Helper()

	query, args, err := sq.
		Select("trip_id", "from_status", "to_status", "reason", "changed_at").
		From("trip_status_history").
		Where(sq.Eq{"trip_id": tripID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	require.NoError(t, err)

	var history trip.TripStatusHistory
	err = pool.QueryRow(ctx, query, args...).Scan(
		&history.TripID,
		&history.FromStatus,
		&history.ToStatus,
		&history.Reason,
		&history.ChangedAt,
	)
	require.NoError(t, err)

	return history
}
