package finish

import (
	"log/slog"
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/response"
	"github.com/StasanDev/avitoLabs/internal/http/trip"
)

type Handler struct {
	tripFinisher tripFinisher
	logger       *slog.Logger
}

func NewHandler(tripFinisher tripFinisher, logger *slog.Logger) *Handler {
	return &Handler{
		tripFinisher: tripFinisher,
		logger:       logger,
	}
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	finishedTrip, err := h.tripFinisher.FinishTrip(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	if err := response.WriteJSON(w, http.StatusOK, trip.ToAPI(finishedTrip)); err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write finish trip response",
			slog.Any("error", err),
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
			"finish trip failed",
			slog.Any("error", err),
		)
	}

	if writeErr != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write finish trip error response",
			slog.Any("error", writeErr),
		)
	}
}
