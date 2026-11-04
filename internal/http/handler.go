package http

import (
	"context"
	"log"
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
}

func NewHandler(
	createHandler *create.Handler,
	getHandler *get.Handler,
	finishHandler *finish.Handler,
	pinger pinger,
	readinessTimeout time.Duration,
) *Handler {
	return &Handler{
		create:           createHandler,
		get:              getHandler,
		finish:           finishHandler,
		pinger:           pinger,
		readinessTimeout: readinessTimeout,
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
	response.WriteJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.readinessTimeout)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		log.Printf("readiness check failed: %v", err)
		response.WriteJSON(w, http.StatusServiceUnavailable, api.HealthResponse{Status: api.Unavailable})
		return
	}

	response.WriteJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}
