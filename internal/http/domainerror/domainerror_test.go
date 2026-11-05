package domainerror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/StasanDev/avitoLabs/internal/http/domainerror"
)

func TestProblem(t *testing.T) {
	internalErr := errors.New("database connection failed: password=secret")

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantTitle  string
		wantDetail string
		wantType   string
	}{
		{
			name:       "validation error",
			err:        fmt.Errorf("validate trip: %w", &trip.ValidationError{Field: "price", Reason: "must be non-negative"}),
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_request",
			wantTitle:  "Invalid request",
			wantDetail: "price: must be non-negative",
			wantType:   "https://tripgo.example/problems/invalid-request",
		},
		{
			name:       "driver busy",
			err:        fmt.Errorf("create trip: %w", trip.ErrDriverBusy),
			wantStatus: http.StatusConflict,
			wantCode:   "driver_busy",
			wantTitle:  "Driver busy",
			wantDetail: "Driver already has an active trip",
			wantType:   "https://tripgo.example/problems/driver-busy",
		},
		{
			name:       "idempotency conflict",
			err:        fmt.Errorf("reserve idempotency key: %w", trip.ErrIdempotencyConflict),
			wantStatus: http.StatusConflict,
			wantCode:   "idempotency_conflict",
			wantTitle:  "Idempotency conflict",
			wantDetail: "Idempotency-Key was already used with a different request body",
			wantType:   "https://tripgo.example/problems/idempotency-conflict",
		},
		{
			name:       "trip not found",
			err:        fmt.Errorf("get trip: %w", trip.ErrTripNotFound),
			wantStatus: http.StatusNotFound,
			wantCode:   "trip_not_found",
			wantTitle:  "Trip not found",
			wantDetail: "no trip with this id",
			wantType:   "https://tripgo.example/problems/trip-not-found",
		},
		{
			name:       "trip completed",
			err:        fmt.Errorf("finish trip: %w", trip.ErrTripCompleted),
			wantStatus: http.StatusConflict,
			wantCode:   "trip_completed",
			wantTitle:  "Trip completed",
			wantDetail: "trip with this id is already done",
			wantType:   "https://tripgo.example/problems/trip-completed",
		},
		{
			name:       "internal error",
			err:        internalErr,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
			wantTitle:  "Internal Server Error",
			wantDetail: "Internal server error",
			wantType:   "https://tripgo.example/problems/internal-error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, problem := domainerror.Problem(tt.err)

			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, int32(tt.wantStatus), problem.Status)
			assert.Equal(t, tt.wantCode, problem.Code)
			assert.Equal(t, tt.wantTitle, problem.Title)
			assert.Equal(t, tt.wantType, problem.Type)
			require.NotNil(t, problem.Detail)
			assert.Equal(t, tt.wantDetail, *problem.Detail)
			assert.Nil(t, problem.Instance)

			if tt.wantStatus == http.StatusInternalServerError {
				assert.NotContains(t, *problem.Detail, internalErr.Error())
				assert.NotContains(t, *problem.Detail, "password")
			}
		})
	}
}
