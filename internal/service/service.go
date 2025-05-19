package service

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"woody_ear/internal/service/converter"
	speachRecognizer "woody_ear/internal/service/speach_recognizer"
	textai "woody_ear/internal/service/text_ai"
	"woody_ear/internal/service/text_ai/gollama"
	"woody_ear/internal/service/text_ai/llops"
)

// Тип AI сервиса
type AIServiceType string

const (
	// LocalLLM использовать локальную LLM модель
	LocalLLM AIServiceType = "local"
	// LLOpsService использовать внешний LLOps сервис
	LLOpsService AIServiceType = "llops"
)

type Service struct {
	recognizer *speachRecognizer.VoskService
	textai     textai.TextAI
	logger     *slog.Logger
}

// ServiceConfig конфигурация сервиса
type ServiceConfig struct {
	// Тип AI сервиса (local или llops)
	AIType AIServiceType
	// Путь к локальной модели (для AIType = local)
	ModelPath string
	// Параметры для LLOps (для AIType = llops)
	LLOpsBaseURL string
	LLOpsToken   string
	LLOpsModelID string
}

func NewService(config ServiceConfig, log *slog.Logger) (*Service, error) {
	recognizer, err := speachRecognizer.NewVoskService()
	if err != nil {
		return nil, fmt.Errorf("can't create recognizer: %w", err)
	}

	var textAIClient textai.TextAI

	switch config.AIType {
	case LocalLLM:
		log.Info("Инициализация локальной LLM модели", "path", config.ModelPath)
		textAIClient, err = gollama.NewLocalLLMClient(config.ModelPath)
		if err != nil {
			return nil, fmt.Errorf("can't create local LLM client: %w", err)
		}
	case LLOpsService:
		log.Info("Инициализация LLOps клиента",
			"baseURL", config.LLOpsBaseURL,
			"modelID", config.LLOpsModelID)
		textAIClient, err = llops.NewLLOpsClient(
			config.LLOpsBaseURL,
			config.LLOpsToken,
			config.LLOpsModelID,
		)
		if err != nil {
			return nil, fmt.Errorf("can't create LLOps client: %w", err)
		}
	default:
		return nil, fmt.Errorf("неизвестный тип AI сервиса: %s", config.AIType)
	}

	return &Service{recognizer: recognizer, textai: textAIClient, logger: log}, nil
}

func (s *Service) HandleFile(header *multipart.FileHeader, tempDir, tempFile string) (string, error) {
	var readyFilePath string
	fileExtension := strings.ToLower(filepath.Ext(header.Filename))

	// Проверяем тип файла
	switch fileExtension {
	case ".mp3":
		// Создаем путь для WAV файла
		wavFilePath := filepath.Join(tempDir, strings.TrimSuffix(filepath.Base(tempFile), fileExtension)+".wav")

		// Конвертируем MP3 в WAV
		var err error
		readyFilePath, err = converter.ConvertMP3ToWavGo(tempFile, wavFilePath)
		if err != nil {
			return "", fmt.Errorf("can't convert mp3 to wav: %w", err)
		}

		// Удаляем временный MP3 файл, так как он больше не нужен
		if err := os.Remove(tempFile); err != nil {
			s.logger.Warn("failed to remove temporary mp3 file", "err", err)
			// Продолжаем выполнение, даже если не удалось удалить временный файл
		}
	case ".wav":
		// Если файл уже в формате WAV, просто используем его
		readyFilePath = tempFile
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileExtension)
	}

	// Распознаем аудио
	recognizeResult, err := s.recognizer.RecognizeAudio(readyFilePath)
	if err != nil {
		return "", fmt.Errorf("can't recognize audio: %w", err)
	}

	// Удаляем временный WAV файл
	if err := os.Remove(readyFilePath); err != nil {
		s.logger.Warn("failed to remove temporary wav file", "err", err)
		// Продолжаем выполнение, даже если не удалось удалить временный файл
	}

	return recognizeResult, nil
}

func (s *Service) HandleTextFromAI(ctx context.Context, text string) (string, error) {
	return s.textai.Analyze(ctx, text)
}
