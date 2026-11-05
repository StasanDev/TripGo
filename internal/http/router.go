package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/middleware"
)

func NewRouter(handler api.ServerInterface) (http.Handler, error) {
	validator, err := middleware.RequestValidator()
	if err != nil {
		return nil, err
	}
	r := chi.NewRouter()

	return api.HandlerWithOptions(handler, api.ChiServerOptions{
		BaseRouter:       r,
		Middlewares:      []api.MiddlewareFunc{validator},
		ErrorHandlerFunc: middleware.ValidationErrorHandler,
	}), nil
}
