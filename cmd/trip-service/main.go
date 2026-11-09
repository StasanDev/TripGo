package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasanDev/avitoLabs/internal/config"
	"github.com/StasanDev/avitoLabs/internal/container"
	"github.com/StasanDev/avitoLabs/internal/infra/telemetry"
)

var serviceVersion = "dev"

func main() {
	if err := run(); err != nil {
		log.Printf("service failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := telemetry.NewLogger(os.Stdout, cfg.LogLevel, cfg.OTel.ServiceName)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	ctx := context.Background()

	app, err := container.BuildApp(ctx, cfg, logger, serviceVersion)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}
	logger.Info("app is built")

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting HTTP server", slog.String("addr", cfg.Http.Addr))
		serverErrors <- app.Run()
	}()

	signalContext, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	var runErr error
	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			runErr = fmt.Errorf("HTTP server: %w", err)
		}
	case <-signalContext.Done():
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	defer cancelShutdown()

	if err := app.Shutdown(shutdownCtx); err != nil {
		return errors.Join(runErr, fmt.Errorf("graceful shutdown: %w", err))
	}

	logger.Info("service stopped")
	return runErr
}
