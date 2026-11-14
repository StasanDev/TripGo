package finish_test

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
	"github.com/StasanDev/avitoLabs/internal/http/trip/finish"
)

type fixture struct {
	service  *MocktripFinisher
	handler  *finish.Handler
	recorder *httptest.ResponseRecorder
}

func TestFinishTripSuccess(t *testing.T) {
	f := newFixture(t)

	tripID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	testTrip := validTrip(t, tripID)
	finishedAt := time.Date(2026, time.November, 5, 11, 0, 0, 0, time.UTC)
	require.NoError(t, testTrip.Finish(finishedAt))

	f.service.EXPECT().FinishTrip(gomock.Any(), tripID).Return(testTrip, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+tripID.String()+"/finish", nil)
	f.handler.FinishTrip(f.recorder, req, tripID)

	require.Equal(t, http.StatusOK, f.recorder.Code)
	assert.Equal(t, "application/json", f.recorder.Header().Get("Content-Type"))

	responseBody := f.recorder.Body.Bytes()
	var response api.Trip
	require.NoError(t, json.Unmarshal(responseBody, &response))
	assert.Equal(t, testTrip.ID, response.Id)
	assert.Equal(t, api.Completed, response.Status)
	require.NotNil(t, response.FinishedAt)
	assert.Equal(t, finishedAt, *response.FinishedAt)
}

func TestFinishTripAlreadyCompleted(t *testing.T) {
	f := newFixture(t)

	tripID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	f.service.EXPECT().FinishTrip(gomock.Any(), tripID).Return(trip.Trip{}, trip.ErrTripCompleted)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+tripID.String()+"/finish", nil)
	f.handler.FinishTrip(f.recorder, req, tripID)

	assertProblem(t, f.recorder, http.StatusConflict, "trip_completed", req.URL.Path)
}

func TestFinishTripInternalError(t *testing.T) {
	f := newFixture(t)

	tripID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	internalErr := errors.New("database password leaked")
	f.service.EXPECT().FinishTrip(gomock.Any(), tripID).Return(trip.Trip{}, internalErr)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+tripID.String()+"/finish", nil)
	f.handler.FinishTrip(f.recorder, req, tripID)

	responseBody := assertProblem(t, f.recorder, http.StatusInternalServerError, "internal_error", req.URL.Path)
	assert.NotContains(t, string(responseBody), internalErr.Error())
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	ctrl := gomock.NewController(t)
	service := NewMocktripFinisher(ctrl)
	logger := slog.New(slog.DiscardHandler)

	return &fixture{
		service:  service,
		handler:  finish.NewHandler(service, logger),
		recorder: httptest.NewRecorder(),
	}
}

func assertProblem(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	wantStatus int,
	wantCode string,
	wantInstance string,
) []byte {
	t.Helper()

	require.Equal(t, wantStatus, recorder.Code)
	assert.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))

	responseBody := recorder.Body.Bytes()
	var problem api.Problem
	require.NoError(t, json.Unmarshal(responseBody, &problem))
	assert.Equal(t, wantCode, problem.Code)
	assert.EqualValues(t, wantStatus, problem.Status)
	require.NotNil(t, problem.Instance)
	assert.Equal(t, wantInstance, *problem.Instance)

	return responseBody
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
