package http

import (
	"net/http"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/middleware"
	"github.com/go-chi/chi/v5"
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
