package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	oapimiddleware "github.com/oapi-codegen/nethttp-middleware"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/domainerror"
	"github.com/StasanDev/avitoLabs/internal/http/response"
)

func RequestValidator(logger *slog.Logger) (func(http.Handler) http.Handler, error) {
	spec, err := api.GetSpec()
	if err != nil {
		return nil, fmt.Errorf("load OpenAPI specification: %w", err)
	}
	spec.Servers = nil
	validationErrorHandler := ValidationErrorHandler(logger)

	return oapimiddleware.OapiRequestValidatorWithOptions(spec, &oapimiddleware.Options{
		Options: openapi3filter.Options{SkipSettingDefaults: true},
		ErrorHandlerWithOpts: func(
			ctx context.Context, err error, w http.ResponseWriter, r *http.Request,
			opts oapimiddleware.ErrorHandlerOpts,
		) {
			if opts.StatusCode == http.StatusInternalServerError {
				logger.ErrorContext(
					ctx,
					"request validation failed",
					slog.Any("error", err),
				)

				_, writeErr := response.WriteError(w, r, err)
				if writeErr != nil {
					logger.ErrorContext(
						ctx,
						"failed to write request validation error response",
						slog.Any("error", writeErr),
					)
				}
				return
			}

			validationErrorHandler(w, r, err)
		},
	}), nil
}

func ValidationErrorHandler(logger *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		writeErr := response.WriteProblem(
			w,
			r,
			http.StatusBadRequest,
			domainerror.InvalidRequest(validationErrorDetail(err)),
		)
		if writeErr != nil {
			logger.ErrorContext(
				r.Context(),
				"failed to write validation error response",
				slog.Any("error", writeErr),
			)
		}
	}
}

func validationErrorDetail(err error) string {
	const fallback = "Request validation failed"

	var schemaErr *openapi3.SchemaError
	if !errors.As(err, &schemaErr) || schemaErr.Reason == "" {
		return fallback
	}

	field := strings.Join(schemaErr.JSONPointer(), ".")
	if field == "" {
		return schemaErr.Reason
	}

	return field + ": " + schemaErr.Reason
}
