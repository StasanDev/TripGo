package get_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/StasanDev/avitoLabs/internal/usecase/trip/get"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetTripSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)

	testTrip := trip.Trip{
		TripInput: trip.TripInput{
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
		ID:         uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		Status:     trip.TripStatusActive,
		StartedAt:  time.Date(2026, time.November, 5, 10, 0, 0, 0, time.UTC),
		FinishedAt: nil,
	}

	tripRepo := NewMocktripRepository(ctrl)
	tripRepo.EXPECT().GetTrip(gomock.Any(), testTrip.ID).Return(testTrip, nil)

	service := get.NewService(tripRepo)
	got, err := service.GetTrip(context.Background(), testTrip.ID)

	require.NoError(t, err)
	assert.Equal(t, testTrip, got)
}

func TestGetTripNotFound(t *testing.T) {
    ctrl := gomock.NewController(t)

    tripRepo := NewMocktripRepository(ctrl)
    tripRepo.EXPECT().GetTrip(gomock.Any(), gomock.Any()).Return(trip.Trip{}, trip.ErrTripNotFound)

    service := get.NewService(tripRepo)
    got, err := service.GetTrip(context.Background(), uuid.New())

    require.ErrorIs(t, err, trip.ErrTripNotFound)
    assert.Equal(t, trip.Trip{}, got)
}

func TestGetTripRepositoryError(t *testing.T) {
    ctrl := gomock.NewController(t)

    repoErr := errors.New("repository error")
    tripRepo := NewMocktripRepository(ctrl)
    tripRepo.EXPECT().GetTrip(gomock.Any(), gomock.Any()).Return(trip.Trip{}, repoErr)

    service := get.NewService(tripRepo)
    got, err := service.GetTrip(context.Background(), uuid.New())

    require.ErrorIs(t, err, repoErr)
    assert.Equal(t, trip.Trip{}, got)
}
