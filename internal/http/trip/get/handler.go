package get

import (
	"log/slog"
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/response"
	"github.com/StasanDev/avitoLabs/internal/http/trip"
)

type Handler struct {
	tripGetter tripGetter
	logger     *slog.Logger
}

func NewHandler(tripGetter tripGetter, logger *slog.Logger) *Handler {
	return &Handler{
		tripGetter: tripGetter,
		logger:     logger,
	}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	foundTrip, err := h.tripGetter.GetTrip(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := response.WriteJSON(w, http.StatusOK, trip.ToAPI(foundTrip)); err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write get trip response",
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
			"get trip failed",
			slog.Any("error", err),
		)
	}

	if writeErr != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to write get trip error response",
			slog.Any("error", writeErr),
		)
	}
}
