package create_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/StasanDev/avitoLabs/internal/usecase/trip/create"
)

const testIdempotencyTTL = 24 * time.Hour

type createTestDependencies struct {
	tripRepository        *MocktripRepository
	historyRepository     *MocktripStatusHistoryRepository
	idempotencyRepository *MockidempotencyRepository
	txManager             *MocktxManager
	service               *create.Service
}

func TestCreateTripSuccess(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInput()

	expectTransaction(deps.txManager)
	savedTrip := expectTripPersistence(t, deps, input)

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	require.NoError(t, err)
	assert.False(t, replayed)
	assert.Equal(t, *savedTrip, createdTrip)
}

func TestCreateTripDriverBusy(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInput()

	expectTransaction(deps.txManager)
	deps.tripRepository.EXPECT().
		CreateTrip(gomock.Any(), gomock.Any()).
		Return(trip.ErrDriverBusy)

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	require.ErrorIs(t, err, trip.ErrDriverBusy)
	assert.Equal(t, trip.Trip{}, createdTrip)
	assert.False(t, replayed)
}

func TestCreateTripInvalidInput(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInput()
	input.UserID = uuid.Nil

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	var validationErr *trip.ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "user_id", validationErr.Field)
	assert.Equal(t, trip.Trip{}, createdTrip)
	assert.False(t, replayed)
}

func TestCreateTripHistoryError(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInput()
	historyErr := errors.New("history repository failed")

	expectTransaction(deps.txManager)
	deps.tripRepository.EXPECT().
		CreateTrip(gomock.Any(), gomock.Any()).
		Return(nil)
	deps.historyRepository.EXPECT().
		AddInTripHistory(gomock.Any(), gomock.Any()).
		Return(historyErr)

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	require.ErrorIs(t, err, historyErr)
	assert.Equal(t, trip.Trip{}, createdTrip)
	assert.False(t, replayed)
}

func TestCreateTripTransactionError(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInput()
	transactionErr := errors.New("transaction failed")

	deps.txManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		Return(transactionErr)

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	require.ErrorIs(t, err, transactionErr)
	assert.Equal(t, trip.Trip{}, createdTrip)
	assert.False(t, replayed)
}

func TestCreateTripWithNewIdempotencyKey(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInputWithIdempotency()

	expectTransaction(deps.txManager)

	var reservedRecord trip.IdempotencyRecord
	deps.idempotencyRepository.EXPECT().
		ReserveOrGet(gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			_ context.Context,
			record trip.IdempotencyRecord,
		) (trip.IdempotencyRecord, bool, error) {
			reservedRecord = record
			return trip.IdempotencyRecord{}, true, nil
		})

	savedTrip := expectTripPersistence(t, deps, input)

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	require.NoError(t, err)
	assert.False(t, replayed)
	assert.Equal(t, *savedTrip, createdTrip)
	assert.Equal(t, *input.IdempotencyKey, reservedRecord.Key)
	assert.Equal(t, input.RequestHash, reservedRecord.RequestHash)
	assert.Equal(t, savedTrip.ID, reservedRecord.TripID)
	assert.Equal(t, savedTrip.StartedAt.Add(testIdempotencyTTL), reservedRecord.ExpiresAt)
}

func TestCreateTripIdempotencyReplay(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInputWithIdempotency()
	replayedTrip := storedTrip(t)
	existingRecord := trip.NewIdempotencyRecord(
		*input.IdempotencyKey,
		input.RequestHash,
		replayedTrip.ID,
		time.Now().Add(time.Hour),
	)

	expectTransaction(deps.txManager)
	deps.idempotencyRepository.EXPECT().
		ReserveOrGet(gomock.Any(), gomock.Any()).
		Return(existingRecord, false, nil)
	deps.tripRepository.EXPECT().
		GetTrip(gomock.Any(), existingRecord.TripID).
		Return(replayedTrip, nil)

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, replayed)
	assert.Equal(t, replayedTrip, createdTrip)
}

func TestCreateTripIdempotencyConflict(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInputWithIdempotency()
	existingRecord := trip.NewIdempotencyRecord(
		*input.IdempotencyKey,
		[]byte("different request hash"),
		uuid.New(),
		time.Now().Add(time.Hour),
	)

	expectTransaction(deps.txManager)
	deps.idempotencyRepository.EXPECT().
		ReserveOrGet(gomock.Any(), gomock.Any()).
		Return(existingRecord, false, nil)

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	require.ErrorIs(t, err, trip.ErrIdempotencyConflict)
	assert.Equal(t, trip.Trip{}, createdTrip)
	assert.False(t, replayed)
}

func TestCreateTripIdempotencyRepositoryError(t *testing.T) {
	deps := newCreateTestDependencies(t)
	input := validTripInputWithIdempotency()
	repositoryErr := errors.New("idempotency repository failed")

	expectTransaction(deps.txManager)
	deps.idempotencyRepository.EXPECT().
		ReserveOrGet(gomock.Any(), gomock.Any()).
		Return(trip.IdempotencyRecord{}, false, repositoryErr)

	createdTrip, replayed, err := deps.service.CreateTrip(context.Background(), input)

	require.ErrorIs(t, err, repositoryErr)
	assert.Equal(t, trip.Trip{}, createdTrip)
	assert.False(t, replayed)
}

func newCreateTestDependencies(t *testing.T) *createTestDependencies {
	t.Helper()

	ctrl := gomock.NewController(t)
	tripRepository := NewMocktripRepository(ctrl)
	historyRepository := NewMocktripStatusHistoryRepository(ctrl)
	idempotencyRepository := NewMockidempotencyRepository(ctrl)
	txManager := NewMocktxManager(ctrl)

	return &createTestDependencies{
		tripRepository:        tripRepository,
		historyRepository:     historyRepository,
		idempotencyRepository: idempotencyRepository,
		txManager:             txManager,
		service: create.NewService(
			tripRepository,
			txManager,
			idempotencyRepository,
			historyRepository,
			testIdempotencyTTL,
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

func expectTripPersistence(
	t *testing.T,
	deps *createTestDependencies,
	input trip.TripInput,
) *trip.Trip {
	t.Helper()

	savedTrip := new(trip.Trip)
	createCall := deps.tripRepository.EXPECT().
		CreateTrip(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, createdTrip trip.Trip) error {
			*savedTrip = createdTrip
			assert.Equal(t, input, createdTrip.TripInput)
			assert.NotEqual(t, uuid.Nil, createdTrip.ID)
			assert.Equal(t, trip.TripStatusActive, createdTrip.Status)
			assert.False(t, createdTrip.StartedAt.IsZero())
			assert.Nil(t, createdTrip.FinishedAt)
			return nil
		})

	historyCall := deps.historyRepository.EXPECT().
		AddInTripHistory(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, history trip.TripStatusHistory) error {
			assert.Equal(t, savedTrip.ID, history.TripID)
			assert.Nil(t, history.FromStatus)
			assert.Equal(t, savedTrip.Status, history.ToStatus)
			require.NotNil(t, history.Reason)
			assert.Equal(t, trip.TripCreatedReason, *history.Reason)
			assert.Equal(t, savedTrip.StartedAt, history.ChangedAt)
			return nil
		})

	gomock.InOrder(createCall, historyCall)

	return savedTrip
}

func validTripInput() trip.TripInput {
	return trip.TripInput{
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
	}
}

func validTripInputWithIdempotency() trip.TripInput {
	input := validTripInput()
	key := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	input.IdempotencyKey = &key
	input.RequestHash = []byte("request hash")
	return input
}

func storedTrip(t *testing.T) trip.Trip {
	t.Helper()

	storedTrip, err := trip.NewTrip(
		validTripInput(),
		uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		time.Date(2026, time.November, 5, 10, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)
	return storedTrip
}
