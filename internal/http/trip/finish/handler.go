package finish

import (
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/response"
	"github.com/StasanDev/avitoLabs/internal/http/trip"
)

type Handler struct {
	tripFinisher tripFinisher
}

func NewHandler(tripFinisher tripFinisher) *Handler {
	return &Handler{tripFinisher: tripFinisher}
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	finishedTrip, err := h.tripFinisher.FinishTrip(r.Context(), id)
	if err != nil {
		response.WriteError(w, r, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, trip.ToAPI(finishedTrip))
}
