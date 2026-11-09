package get_test

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/trip/get"
)

type fixture struct {
	service  *MocktripGetter
	handler  *get.Handler
	recorder *httptest.ResponseRecorder
}

func TestGetTripSuccess(t *testing.T) {
	f := newFixture(t)

	tripID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	testTrip := validTrip(t, tripID)

	f.service.EXPECT().GetTrip(gomock.Any(), tripID).Return(testTrip, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+tripID.String(), nil)
	f.handler.GetTrip(f.recorder, req, tripID)

	require.Equal(t, http.StatusOK, f.recorder.Code)
	assert.Equal(t, "application/json", f.recorder.Header().Get("Content-Type"))

	var response api.Trip
	require.NoError(t, json.NewDecoder(f.recorder.Body).Decode(&response))
	assert.Equal(t, testTrip.ID, response.Id)
	assert.Equal(t, api.Active, response.Status)
	assert.Equal(t, testTrip.UserID, response.UserId)
	assert.Equal(t, testTrip.DriverID, response.DriverId)
	assert.Equal(t, testTrip.StartPoint.Latitude, response.StartPoint.Latitude)
	assert.Equal(t, testTrip.StartPoint.Longitude, response.StartPoint.Longitude)
	assert.Equal(t, testTrip.EndPoint.Latitude, response.EndPoint.Latitude)
	assert.Equal(t, testTrip.EndPoint.Longitude, response.EndPoint.Longitude)
	assert.Equal(t, testTrip.Price, response.Price)
	assert.Equal(t, testTrip.StartedAt, response.StartedAt)
	assert.Nil(t, response.FinishedAt)
}

func TestGetTripNotFound(t *testing.T) {
	f := newFixture(t)

	tripID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	f.service.EXPECT().GetTrip(gomock.Any(), tripID).Return(trip.Trip{}, trip.ErrTripNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+tripID.String(), nil)
	f.handler.GetTrip(f.recorder, req, tripID)

	require.Equal(t, http.StatusNotFound, f.recorder.Code)
	assert.Equal(
		t,
		"application/problem+json",
		f.recorder.Header().Get("Content-Type"),
	)

	var problem api.Problem
	require.NoError(t, json.NewDecoder(f.recorder.Body).Decode(&problem))
	assert.Equal(t, "trip_not_found", problem.Code)
	assert.EqualValues(t, http.StatusNotFound, problem.Status)
	require.NotNil(t, problem.Instance)
	assert.Equal(t, "/api/v1/trips/"+tripID.String(), *problem.Instance)
}

func TestGetTripInternalError(t *testing.T) {
	f := newFixture(t)

	tripID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	internalError := errors.New("database password leaked")
	f.service.EXPECT().GetTrip(gomock.Any(), tripID).Return(trip.Trip{}, internalError)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+tripID.String(), nil)
	f.handler.GetTrip(f.recorder, req, tripID)

	require.Equal(t, http.StatusInternalServerError, f.recorder.Code)
	assert.Equal(
		t,
		"application/problem+json",
		f.recorder.Header().Get("Content-Type"),
	)
	responseBody := f.recorder.Body.Bytes()
	var problem api.Problem
	require.NoError(t, json.Unmarshal(responseBody, &problem))
	assert.Equal(t, "internal_error", problem.Code)
	assert.EqualValues(t, http.StatusInternalServerError, problem.Status)
	require.NotNil(t, problem.Instance)
	assert.Equal(t, "/api/v1/trips/"+tripID.String(), *problem.Instance)
	assert.NotContains(t, string(responseBody), internalError.Error())
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	ctrl := gomock.NewController(t)
	service := NewMocktripGetter(ctrl)
	logger := slog.New(slog.DiscardHandler)

	return &fixture{
		service:  service,
		handler:  get.NewHandler(service, logger),
		recorder: httptest.NewRecorder(),
	}
}

func validTrip(t *testing.T, id uuid.UUID) trip.Trip {
	t.Helper()

	result, err := trip.NewTrip(
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
		id,
		time.Date(2026, time.November, 5, 10, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)

	return result
}
