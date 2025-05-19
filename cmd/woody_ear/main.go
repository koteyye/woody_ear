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

// Значения по умолчанию
const (
	defaultBaseURL      = "localhost:8080"
	defaultLLMModelPath = "./models/llama-2-7b-chat.gguf"
	defaultAIType       = "local" // По умолчанию используем локальную модель
	defaultLLOpsBaseURL = "http://localhost:11434"
	defaultLLOpsModelID = "llama2"
)

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Получаем значения из переменных окружения или используем значения по умолчанию
	baseURL := getEnvOrDefault("SERVER_ADDRESS", defaultBaseURL)
	llmModelPath := getEnvOrDefault("LLM_MODEL_PATH", defaultLLMModelPath)
	aiType := getEnvOrDefault("AI_TYPE", defaultAIType)
	llopsBaseURL := getEnvOrDefault("LLOPS_BASE_URL", defaultLLOpsBaseURL)
	llopsToken := getEnvOrDefault("LLOPS_TOKEN", "")
	llopsModelID := getEnvOrDefault("LLOPS_MODEL_ID", defaultLLOpsModelID)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	// Настраиваем логгер
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = slog.LevelDebug
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	})

	logger := slog.New(jsonHandler)
	logger.Info("Starting application", "version", "1.0.0")

	defer func() {
		if msg := recover(); msg != nil {
			logger.Error("panic", "msg", msg)
			cancel()
		}
	}()

	// Создаем конфигурацию сервиса
	serviceConfig := service.ServiceConfig{
		AIType:       service.AIServiceType(aiType),
		ModelPath:    llmModelPath,
		LLOpsBaseURL: llopsBaseURL,
		LLOpsToken:   llopsToken,
		LLOpsModelID: llopsModelID,
	}

	// Инициализируем сервис
	logger.Info("Initializing service", "ai_type", aiType)
	service, err := service.NewService(serviceConfig, logger)
	if err != nil {
		logger.Error("failed to create service", "err", err)
		os.Exit(1)
	}
	httpServer := restapi.NewHTTPServer(baseURL, logger, service)
	httpRouter, err := httpServer.NewRouter()
	if err != nil {
		logger.Error("failed to create HTTP router", "err", err)
		os.Exit(1)
	}

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		logger.Info("Starting server", "url", fmt.Sprintf("http://%s", baseURL))

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
