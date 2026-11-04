package create

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/response"
	"github.com/StasanDev/avitoLabs/internal/http/trip"
)

type Handler struct {
	tripCreator tripCreator
}

func NewHandler(tripCreator tripCreator) *Handler {
	return &Handler{tripCreator: tripCreator}
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteInvalidRequest(w, r, err)
		return
	}

	var body api.TripData
	if err := json.Unmarshal(payload, &body); err != nil {
		response.WriteInvalidRequest(w, r, err)
		return
	}
	if err := validateTripData(body); err != nil {
		response.WriteInvalidRequest(w, r, err)
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
		response.WriteError(w, r, err)
		return
	}
	if replayed {
		response.WriteJSON(w, http.StatusOK, trip.ToAPI(createdTrip))
		return
	}

	w.Header().Set("Location", "/api/v1/trips/"+createdTrip.ID.String())
	response.WriteJSON(w, http.StatusCreated, trip.ToAPI(createdTrip))
}
