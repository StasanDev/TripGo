package get

import (
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/response"
	"github.com/StasanDev/avitoLabs/internal/http/trip"
)

type Handler struct {
	tripGetter tripGetter
}

func NewHandler(tripGetter tripGetter) *Handler {
	return &Handler{tripGetter: tripGetter}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	foundTrip, err := h.tripGetter.GetTrip(r.Context(), id)
	if err != nil {
		response.WriteError(w, r, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, trip.ToAPI(foundTrip))
}
