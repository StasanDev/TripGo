package domainerror

import (
	"errors"
	"net/http"
	"strings"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	api "github.com/StasanDev/avitoLabs/internal/generated"
)

func Problem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, trip.ErrInvalidInput):
		return http.StatusBadRequest, InvalidRequest(err.Error())
	case errors.Is(err, trip.ErrDriverBusy):
		return http.StatusConflict, problem(
			http.StatusConflict, "driver_busy", "Driver busy", "Driver already has an active trip",
		)
	case errors.Is(err, trip.ErrIdempotencyConflict):
		return http.StatusConflict, problem(
			http.StatusConflict, "idempotency_conflict", "Idempotency conflict",
			"Idempotency-Key was already used with a different request body",
		)
	case errors.Is(err, trip.ErrTripNotFound):
		return http.StatusNotFound, problem(
			http.StatusNotFound, "trip_not_found", "Trip not found", "no trip with this id",
		)
	case errors.Is(err, trip.ErrTripCompleted):
		return http.StatusConflict, problem(
			http.StatusConflict, "trip_completed", "Trip completed", "trip with this id is already done",
		)
	default:
		return http.StatusInternalServerError, problem(
			http.StatusInternalServerError, "internal_error", "Internal Server Error", "Internal server error",
		)
	}
}

func InvalidRequest(detail string) api.Problem {
	return problem(http.StatusBadRequest, "invalid_request", "Invalid request", detail)
}

func problem(status int, code, title, detail string) api.Problem {
	return api.Problem{
		Type:   "https://tripgo.example/problems/" + strings.ReplaceAll(code, "_", "-"),
		Title:  title,
		Status: int32(status),
		Detail: &detail,
		Code:   code,
	}
}
