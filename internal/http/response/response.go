package response

import (
	"encoding/json"
	"log"
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/domainerror"
)

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode JSON response: %v", err)
	}
}

func WriteProblem(w http.ResponseWriter, r *http.Request, status int, problem api.Problem) {
	instance := r.URL.Path
	problem.Instance = &instance

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(problem); err != nil {
		log.Printf("encode problem response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status, problem := domainerror.Problem(err)
	if status == http.StatusInternalServerError {
		log.Printf("request failed: %v", err)
	}
	WriteProblem(w, r, status, problem)
}

func WriteInvalidRequest(w http.ResponseWriter, r *http.Request, err error) {
	WriteProblem(w, r, http.StatusBadRequest, domainerror.InvalidRequest(err.Error()))
}
