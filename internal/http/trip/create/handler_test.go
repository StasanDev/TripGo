package create_test

import (
	"bytes"
	"crypto/sha256"
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
	"github.com/StasanDev/avitoLabs/internal/http/trip/create"
)

const validRequestBody = `{
	"user_id": "11111111-1111-4111-8111-111111111111",
	"driver_id": "22222222-2222-4222-8222-222222222222",
	"start_point": {
		"latitude": 55.7558,
		"longitude": 37.6173
	},
	"end_point": {
		"latitude": 59.9343,
		"longitude": 30.3351
	},
	"price": 1500
}`

type fixture struct {
	service  *MocktripCreator
	handler  *create.Handler
	recorder *httptest.ResponseRecorder
}

func TestCreateTripSuccess(t *testing.T) {
	f := newFixture(t)

	input := validTripInput()
	createdTrip := testTrip(t, input)

	f.service.EXPECT().CreateTrip(gomock.Any(), input).Return(createdTrip, false, nil)

	req := newCreateTripRequest([]byte(validRequestBody))

	f.handler.CreateTrip(f.recorder, req, api.CreateTripParams{})

	require.Equal(t, http.StatusCreated, f.recorder.Code)
	assert.Equal(t, "application/json", f.recorder.Header().Get("Content-Type"))
	assert.Equal(t, "/api/v1/trips/"+createdTrip.ID.String(), f.recorder.Header().Get("Location"))
	assertTripResponse(t, f.recorder.Body.Bytes(), createdTrip)
}

func TestCreateTripIdempotencyReplay(t *testing.T) {
	f := newFixture(t)

	key := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	requestBody := []byte(validRequestBody)
	input := tripInputWithIdempotency(key, requestBody)
	existingTrip := testTrip(t, validTripInput())

	f.service.EXPECT().CreateTrip(gomock.Any(), input).Return(existingTrip, true, nil)

	req := newCreateTripRequest(requestBody)
	f.handler.CreateTrip(f.recorder, req, api.CreateTripParams{IdempotencyKey: &key})

	require.Equal(t, http.StatusOK, f.recorder.Code)
	assert.Equal(t, "application/json", f.recorder.Header().Get("Content-Type"))
	assert.Empty(t, f.recorder.Header().Get("Location"))
	assertTripResponse(t, f.recorder.Body.Bytes(), existingTrip)
}

func TestCreateTripPassesIdempotencyKeyAndRequestHash(t *testing.T) {
	f := newFixture(t)
	key := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	requestBody := []byte(validRequestBody)
	expectedInput := tripInputWithIdempotency(key, requestBody)
	createdTrip := testTrip(t, validTripInput())

	f.service.EXPECT().CreateTrip(gomock.Any(), expectedInput).Return(createdTrip, false, nil)

	req := newCreateTripRequest(requestBody)
	f.handler.CreateTrip(f.recorder, req, api.CreateTripParams{IdempotencyKey: &key})

	require.Equal(t, http.StatusCreated, f.recorder.Code)
}

func TestCreateTripInvalidJSON(t *testing.T) {
	f := newFixture(t)
	req := newCreateTripRequest([]byte(`{"user_id":`))

	f.handler.CreateTrip(f.recorder, req, api.CreateTripParams{})

	assertProblem(t, f.recorder, http.StatusBadRequest, "invalid_request", req.URL.Path)
}

func TestCreateTripDriverBusy(t *testing.T) {
	f := newFixture(t)
	input := validTripInput()

	f.service.EXPECT().CreateTrip(gomock.Any(), input).Return(trip.Trip{}, false, trip.ErrDriverBusy)

	req := newCreateTripRequest([]byte(validRequestBody))

	f.handler.CreateTrip(f.recorder, req, api.CreateTripParams{})

	assertProblem(t, f.recorder, http.StatusConflict, "driver_busy", req.URL.Path)
}

func TestCreateTripInternalError(t *testing.T) {
	f := newFixture(t)
	input := validTripInput()
	internalErr := errors.New("database password leaked")

	f.service.EXPECT().CreateTrip(gomock.Any(), input).Return(trip.Trip{}, false, internalErr)

	req := newCreateTripRequest([]byte(validRequestBody))
	f.handler.CreateTrip(f.recorder, req, api.CreateTripParams{})

	responseBody := assertProblem(t, f.recorder, http.StatusInternalServerError, "internal_error", req.URL.Path)
	assert.NotContains(t, string(responseBody), internalErr.Error())
}

func TestCreateTripIdempotencyConflict(t *testing.T) {
	f := newFixture(t)
	key := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	requestBody := []byte(validRequestBody)
	input := tripInputWithIdempotency(key, requestBody)

	f.service.EXPECT().CreateTrip(gomock.Any(), input).Return(trip.Trip{}, false, trip.ErrIdempotencyConflict)

	req := newCreateTripRequest(requestBody)
	f.handler.CreateTrip(f.recorder, req, api.CreateTripParams{IdempotencyKey: &key})

	assertProblem(t, f.recorder, http.StatusConflict, "idempotency_conflict", req.URL.Path)
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	ctrl := gomock.NewController(t)
	service := NewMocktripCreator(ctrl)
	logger := slog.New(slog.DiscardHandler)

	return &fixture{
		service:  service,
		handler:  create.NewHandler(service, logger),
		recorder: httptest.NewRecorder(),
	}
}

func newCreateTripRequest(body []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
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

func tripInputWithIdempotency(key uuid.UUID, requestBody []byte) trip.TripInput {
	input := validTripInput()
	requestHash := sha256.Sum256(requestBody)
	input.IdempotencyKey = &key
	input.RequestHash = requestHash[:]
	return input
}

func testTrip(t *testing.T, input trip.TripInput) trip.Trip {
	t.Helper()

	createdTrip, err := trip.NewTrip(
		input,
		uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		time.Date(2026, time.November, 5, 10, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)
	return createdTrip
}

func assertTripResponse(t *testing.T, responseBody []byte, expected trip.Trip) {
	t.Helper()

	var response api.Trip
	require.NoError(t, json.Unmarshal(responseBody, &response))
	assert.Equal(t, expected.ID, response.Id)
	assert.Equal(t, expected.UserID, response.UserId)
	assert.Equal(t, expected.DriverID, response.DriverId)
	assert.Equal(t, expected.StartPoint.Latitude, response.StartPoint.Latitude)
	assert.Equal(t, expected.StartPoint.Longitude, response.StartPoint.Longitude)
	assert.Equal(t, expected.EndPoint.Latitude, response.EndPoint.Latitude)
	assert.Equal(t, expected.EndPoint.Longitude, response.EndPoint.Longitude)
	assert.Equal(t, expected.Price, response.Price)
	assert.Equal(t, api.TripStatus(expected.Status), response.Status)
	assert.Equal(t, expected.StartedAt, response.StartedAt)
	assert.Equal(t, expected.FinishedAt, response.FinishedAt)
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
