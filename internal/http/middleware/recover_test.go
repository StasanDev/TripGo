package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/middleware"
)

func TestRecover(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := middleware.Recover(logger)(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {
			panic("sensitive panic text")
		},
	))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)

	require.NotPanics(t, func() {
		handler.ServeHTTP(recorder, request)
	})

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))

	var problem api.Problem
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&problem))
	assert.Equal(t, "internal_error", problem.Code)
	assert.NotContains(t, recorder.Body.String(), "sensitive panic text")

	var logRecord map[string]any
	require.NoError(t, json.NewDecoder(&logs).Decode(&logRecord))
	assert.Equal(t, "ERROR", logRecord["level"])
	assert.Equal(t, "panic recovered", logRecord["msg"])
	assert.Equal(t, "sensitive panic text", logRecord["panic"])
	assert.Contains(t, logRecord["stack"], "TestRecover")
}
