package restapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"woody_ear/internal/service"

	"github.com/go-chi/chi/v5"
)

const (
	readHeaderTimeout = 5 * time.Minute
)

type httpServer struct {
	address string
	log     *slog.Logger
	service *service.Service
}

func NewHTTPServer(address string, log *slog.Logger, service *service.Service) *httpServer {
	return &httpServer{
		address: address,
		log:     log,
		service: service,
	}
}

func (s *httpServer) NewRouter() (*chi.Mux, error) {
	router := chi.NewRouter()

	// Middleware для CORS
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// Обработчик для проверки здоровья приложения
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Woody Ear Service is running"))
	})

	// Обработчик для загрузки файлов
	router.Post("/upload", s.uploadFile)

	// Обработчик для статических файлов (HTML, CSS, JS)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		// Сначала ищем файл в корне проекта
		if _, err := os.Stat("test-upload.html"); err == nil {
			http.ServeFile(w, r, "test-upload.html")
			return
		}
		
		// Если файл не найден в корне, ищем в директории scripts
		if _, err := os.Stat("scripts/test-upload.html"); err == nil {
			http.ServeFile(w, r, "scripts/test-upload.html")
			return
		}
		
		// Если файл не найден, возвращаем ошибку
		http.Error(w, "Файл test-upload.html не найден", http.StatusNotFound)
	})

	return router, nil
}

func (s *httpServer) Start(ctx context.Context, router *chi.Mux) error {
	server := &http.Server{
		Addr:        s.address,
		Handler:     router,
		ReadTimeout: readHeaderTimeout,
		WriteTimeout: readHeaderTimeout,
	}

	go func() {
		<-ctx.Done()
		err := server.Shutdown(context.Background())
		if err != nil {
			s.log.Error("server shutdown with error", "err", err)
		}
	}()

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error("server listen and serve with error", "err", err)
	}
	return nil
}

const (
	maxUploadSize = 50 * 1024 * 1024 // 50 MB
)

func (s *httpServer) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем размер файла
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		s.log.Error("ошибка при парсинге формы", "err", err)
		http.Error(w, "Файл слишком большой (максимум 50MB)", http.StatusBadRequest)
		return
	}

	// Читаем файл из запроса
	file, header, err := r.FormFile("file")
	if err != nil {
		s.log.Error("ошибка при чтении файла", "err", err)
		http.Error(w, "Ошибка при чтении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Проверяем тип файла
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	if fileExt != ".mp3" && fileExt != ".wav" {
		s.log.Error("неподдерживаемый тип файла", "type", fileExt)
		http.Error(w, "Поддерживаются только файлы MP3 и WAV", http.StatusBadRequest)
		return
	}

	// Создаем временный файл с уникальным именем
	tempDir := os.TempDir()
	tempFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	tempFile := filepath.Join(tempDir, tempFileName)
	out, err := os.Create(tempFile)
	if err != nil {
		s.log.Error("ошибка при создании временного файла", "err", err)
		http.Error(w, "Ошибка при создании файла", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Копируем данные во временный файл
	if _, err = io.Copy(out, file); err != nil {
		s.log.Error("ошибка при сохранении файла", "err", err)
		http.Error(w, "Ошибка при сохранении файла", http.StatusInternalServerError)
		return
	}

	// Закрываем файл перед обработкой
	out.Close()

	// Обрабатываем файл
	s.log.Info("начинаем обработку файла", "filename", header.Filename, "size", header.Size)
	recognizeResult, err := s.service.HandleFile(header, tempDir, tempFile)
	if err != nil {
		s.log.Error("ошибка при обработке файла", "err", err)
		http.Error(w, "Ошибка при обработке файла: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обрабатываем распознанный текст через AI
	s.log.Info("начинаем обработку текста через AI", "text_length", len(recognizeResult))
	aiTextResult, err := s.service.HandleTextFromAI(context.Background(), recognizeResult)
	if err != nil {
		s.log.Error("ошибка при обработке текста от AI", "err", err)
		http.Error(w, "Ошибка при обработке текста от AI: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	body, err := json.Marshal(response{
		Reconize:     recognizeResult,
		AIOutputText: aiTextResult,
	})
	if err != nil {
		s.log.Error("ошибка при marshalling ответа", "err", err)
		http.Error(w, "Ошибка при формировании ответа", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
