package create

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/response"
	"github.com/StasanDev/avitoLabs/internal/http/trip"
)

type Handler struct {
	tripCreator tripCreator
	logger      *slog.Logger
}

func NewHandler(tripCreator tripCreator, logger *slog.Logger) *Handler {
	return &Handler{
		tripCreator: tripCreator,
		logger:      logger,
	}
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeInvalidRequest(w, r, err)
		return
	}

	var body api.TripData
	if err := json.Unmarshal(payload, &body); err != nil {
		h.writeInvalidRequest(w, r, err)
		return
	}

	input := tripInputFromAPI(body)
	if params.IdempotencyKey != nil {
		hash := sha256.Sum256(payload)
		input.IdempotencyKey = params.IdempotencyKey
		input.RequestHash = hash[:]
	}

	createdTrip, replayed, err := h.tripCreator.CreateTrip(r.Context(), input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	status := http.StatusCreated
	if replayed {
		status = http.StatusOK
	} else {
		w.Header().Set("Location", "/api/v1/trips/"+createdTrip.ID.String())
	}

	if err := response.WriteJSON(w, status, trip.ToAPI(createdTrip)); err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write create trip response",
			slog.Any("error", err),
		)
	}
}

func (h *Handler) writeInvalidRequest(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if writeErr := response.WriteInvalidRequest(w, r, err); writeErr != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write create trip invalid request response",
			slog.Any("error", writeErr),
		)
	}
}

func (h *Handler) writeError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	status, writeErr := response.WriteError(w, r, err)

	if status == http.StatusInternalServerError {
		h.logger.ErrorContext(
			r.Context(),
			"create trip failed",
			slog.Any("error", err),
		)
	}

	if writeErr != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write create trip error response",
			slog.Any("error", writeErr),
		)
	}
}
