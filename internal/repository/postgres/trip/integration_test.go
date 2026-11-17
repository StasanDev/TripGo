//go:build integration

package repository_test

import (
	"context"
	"errors"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func (s *RepositorySuite) TestInsertInTwoTablesRollback() {
	t := s.T()
	testTrip := newTestTrip(t)
	history := trip.NewTripStatusHistory(
		testTrip,
		nil,
		trip.TripCreatedReason,
		testTrip.StartedAt,
	)
	rollbackErr := errors.New("force rollback")

	err := s.infra.TxManager.Do(t.Context(), func(ctx context.Context) error {
		if err := s.tripRepo.CreateTrip(ctx, testTrip); err != nil {
			return err
		}
		if err := s.tripStatusHistoryRepository.AddInTripHistory(ctx, history); err != nil {
			return err
		}
		return rollbackErr
	})
	s.Require().ErrorIs(err, rollbackErr)

	got, err := s.tripRepo.GetTrip(t.Context(), testTrip.ID)
	s.Require().ErrorIs(err, trip.ErrTripNotFound)
	s.Empty(got)

	var historyCount int
	err = s.infra.Pool.QueryRow(
		t.Context(),
		`SELECT COUNT(*) FROM trip_status_history WHERE trip_id = $1`,
		testTrip.ID,
	).Scan(&historyCount)
	s.Require().NoError(err)
	s.Zero(historyCount)
}
