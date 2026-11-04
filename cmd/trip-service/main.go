package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasanDev/avitoLabs/internal/config"
	"github.com/StasanDev/avitoLabs/internal/container"
)

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
	// log.Println(cfg)
	log.Printf("config is loaded")

	ctx := context.Background()
	app, err := container.BuildApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}
	log.Printf("app is built")

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("starting HTTP server on %s", cfg.Http.Addr)
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

	log.Printf("service stopped")
	return runErr
}
