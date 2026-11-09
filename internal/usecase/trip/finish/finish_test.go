package finish_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/StasanDev/avitoLabs/internal/usecase/trip/finish"
)

type finishTestDependencies struct {
	tripRepository    *MocktripRepository
	historyRepository *MocktripStatusHistoryRepository
	txManager         *MocktxManager
	service           *finish.Service
}

func TestFinishTripSuccess(t *testing.T) {
	deps := newFinishTestDependencies(t)
	activeTrip := storedActiveTrip(t)

	expectTransaction(deps.txManager)
	getCall := deps.tripRepository.EXPECT().
		GetTrip(gomock.Any(), activeTrip.ID).
		Return(activeTrip, nil)

	var persistedTrip trip.Trip
	finishCall := deps.tripRepository.EXPECT().
		FinishTrip(gomock.Any(), activeTrip.ID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, finishedAt time.Time) (trip.Trip, error) {
			persistedTrip = activeTrip
			require.NoError(t, persistedTrip.Finish(finishedAt))
			return persistedTrip, nil
		})

	historyCall := deps.historyRepository.EXPECT().
		AddInTripHistory(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, history trip.TripStatusHistory) error {
			assert.Equal(t, persistedTrip.ID, history.TripID)
			require.NotNil(t, history.FromStatus)
			assert.Equal(t, trip.TripStatusActive, *history.FromStatus)
			assert.Equal(t, trip.TripStatusCompleted, history.ToStatus)
			require.NotNil(t, history.Reason)
			assert.Equal(t, trip.TripFinishedReason, *history.Reason)
			require.NotNil(t, persistedTrip.FinishedAt)
			assert.Equal(t, *persistedTrip.FinishedAt, history.ChangedAt)
			return nil
		})

	gomock.InOrder(getCall, finishCall, historyCall)

	finishedTrip, err := deps.service.FinishTrip(context.Background(), activeTrip.ID)

	require.NoError(t, err)
	assert.Equal(t, persistedTrip, finishedTrip)
	assert.Equal(t, trip.TripStatusCompleted, finishedTrip.Status)
	require.NotNil(t, finishedTrip.FinishedAt)
}

func TestFinishTripAlreadyCompleted(t *testing.T) {
	deps := newFinishTestDependencies(t)
	completedTrip := storedCompletedTrip(t)

	expectTransaction(deps.txManager)
	deps.tripRepository.EXPECT().
		GetTrip(gomock.Any(), completedTrip.ID).
		Return(completedTrip, nil)

	finishedTrip, err := deps.service.FinishTrip(context.Background(), completedTrip.ID)

	require.ErrorIs(t, err, trip.ErrTripCompleted)
	assert.Equal(t, trip.Trip{}, finishedTrip)
}

func TestFinishTripNotFound(t *testing.T) {
	deps := newFinishTestDependencies(t)
	tripID := uuid.New()

	expectTransaction(deps.txManager)
	deps.tripRepository.EXPECT().
		GetTrip(gomock.Any(), tripID).
		Return(trip.Trip{}, trip.ErrTripNotFound)

	finishedTrip, err := deps.service.FinishTrip(context.Background(), tripID)

	require.ErrorIs(t, err, trip.ErrTripNotFound)
	assert.Equal(t, trip.Trip{}, finishedTrip)
}

func newFinishTestDependencies(t *testing.T) *finishTestDependencies {
	t.Helper()

	ctrl := gomock.NewController(t)
	tripRepository := NewMocktripRepository(ctrl)
	historyRepository := NewMocktripStatusHistoryRepository(ctrl)
	txManager := NewMocktxManager(ctrl)

	return &finishTestDependencies{
		tripRepository:    tripRepository,
		historyRepository: historyRepository,
		txManager:         txManager,
		service: finish.NewService(
			tripRepository,
			historyRepository,
			txManager,
			nil,
		),
	}
}

func expectTransaction(txManager *MocktxManager) {
	txManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

func storedActiveTrip(t *testing.T) trip.Trip {
	t.Helper()

	storedTrip, err := trip.NewTrip(
		trip.TripInput{
			UserID:   uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			DriverID: uuid.MustParse("22222222-2222-4222-8222-222222222222"),
			StartPoint: trip.Coordinates{
				Latitude:  55.7558,
				Longitude: 37.6173,
			},
			EndPoint: trip.Coordinates{
				Latitude:  59.9343,
				Longitude: 30.3351,
			},
			Price: 1500,
		},
		uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		time.Date(2026, time.November, 5, 10, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)
	return storedTrip
}

func storedCompletedTrip(t *testing.T) trip.Trip {
	t.Helper()

	storedTrip := storedActiveTrip(t)
	finishedAt := time.Date(2026, time.November, 5, 11, 0, 0, 0, time.UTC)
	require.NoError(t, storedTrip.Finish(finishedAt))
	return storedTrip
}
