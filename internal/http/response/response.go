package response

import (
	"encoding/json"
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/domainerror"
)

func WriteJSON(w http.ResponseWriter, status int, value any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}

func WriteProblem(w http.ResponseWriter, r *http.Request, status int, problem api.Problem) error {
	instance := r.URL.Path
	problem.Instance = &instance

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(problem)
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) (int, error) {
	status, problem := domainerror.Problem(err)
	return status, WriteProblem(w, r, status, problem)
}

func WriteInvalidRequest(w http.ResponseWriter, r *http.Request, err error) error {
	return WriteProblem(w, r, http.StatusBadRequest, domainerror.InvalidRequest(err.Error()))
}
