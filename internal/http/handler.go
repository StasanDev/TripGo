package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/response"
	"github.com/StasanDev/avitoLabs/internal/http/trip/create"
	"github.com/StasanDev/avitoLabs/internal/http/trip/finish"
	"github.com/StasanDev/avitoLabs/internal/http/trip/get"
)

type Handler struct {
	create           *create.Handler
	get              *get.Handler
	finish           *finish.Handler
	pinger           pinger
	readinessTimeout time.Duration
	logger           *slog.Logger
}

func NewHandler(
	createHandler *create.Handler,
	getHandler *get.Handler,
	finishHandler *finish.Handler,
	pinger pinger,
	readinessTimeout time.Duration,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		create:           createHandler,
		get:              getHandler,
		finish:           finishHandler,
		pinger:           pinger,
		readinessTimeout: readinessTimeout,
		logger:           logger,
	}
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	h.create.CreateTrip(w, r, params)
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	h.get.GetTrip(w, r, id)
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	h.finish.FinishTrip(w, r, id)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if err := response.WriteJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok}); err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write health response",
			slog.Any("error", err),
		)
	}
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.readinessTimeout)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"readiness check failed",
			slog.Any("error", err),
		)
		if err := response.WriteJSON(
			w,
			http.StatusServiceUnavailable,
			api.HealthResponse{Status: api.Unavailable},
		); err != nil {
			h.logger.ErrorContext(
				r.Context(),
				"failed to write readiness response",
				slog.Any("error", err),
			)
		}
		return
	}

	if err := response.WriteJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok}); err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write readiness response",
			slog.Any("error", err),
		)
	}
}
