package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	oapimiddleware "github.com/oapi-codegen/nethttp-middleware"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/domainerror"
	"github.com/StasanDev/avitoLabs/internal/http/response"
)

func RequestValidator() (func(http.Handler) http.Handler, error) {
	spec, err := api.GetSpec()
	if err != nil {
		return nil, fmt.Errorf("load OpenAPI specification: %w", err)
	}
	spec.Servers = nil

	return oapimiddleware.OapiRequestValidatorWithOptions(spec, &oapimiddleware.Options{
		Options: openapi3filter.Options{SkipSettingDefaults: true},
		ErrorHandlerWithOpts: func(
			_ context.Context, err error, w http.ResponseWriter, r *http.Request,
			opts oapimiddleware.ErrorHandlerOpts,
		) {
			if opts.StatusCode == http.StatusInternalServerError {
				response.WriteError(w, r, err)
				return
			}
			ValidationErrorHandler(w, r, err)
		},
	}), nil
}

func ValidationErrorHandler(w http.ResponseWriter, r *http.Request, _ error) {
	response.WriteProblem(w, r, http.StatusBadRequest, domainerror.InvalidRequest("Request validation failed"))
}
