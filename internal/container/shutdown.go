package container

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StasanDev/avitoLabs/internal/infra/telemetry"
)

func (a *App) Shutdown(ctx context.Context) error {
	var result error

	if err := a.server.Shutdown(ctx); err != nil {
		result = errors.Join(result, fmt.Errorf("shutdown HTTP server: %w", err))
	}

	a.pool.Close()

	return errors.Join(
		result,
		shutdownTelemetryProviders(a.telemetryProviders, a.shutdownTimeout),
	)
}

func shutdownTelemetryProviders(providers *telemetry.Providers, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := providers.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown telemetry: %w", err)
	}

	return nil
}
