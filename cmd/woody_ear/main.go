package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	restapi "woody_ear/internal/api/http"
	"woody_ear/internal/service"

	"golang.org/x/sync/errgroup"
)

const baseURL = "localhost:8080"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})

	logger := slog.New(jsonHandler)

	defer func() {
		if msg := recover(); msg != nil {
			logger.Error("panic", "msg", msg)
			cancel()
		}
	}()

	service, err := service.NewService()
	httpServer := restapi.NewHTTPServer(baseURL, logger, service)
	httpRouter, err := httpServer.NewRouter()
	if err != nil {
		logger.Error("failed to create HTTP router", "err", err)
		os.Exit(1)
	}

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		logger.Info("Starting server", "url", "http://localhost:8080")

		errN := httpServer.Start(gCtx, httpRouter)
		if errN != nil {
			return fmt.Errorf("error on listen http server: %w", errN)
		}
		logger.Info("graceful http server stop")
		return nil
	})

	if err := group.Wait(); err != nil {
		logger.Error("group wait error", "err", err)
		os.Exit(1)
	}
}
