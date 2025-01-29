package restapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"woody_ear/internal/service"

	"github.com/go-chi/chi/v5"
)

const (
	readHeaderTimeout = 15 * time.Second
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

	router.Post("/upload", s.uploadFile)

	return router, nil
}

func (s *httpServer) Start(ctx context.Context, router *chi.Mux) error {
	server := &http.Server{
		Addr:        s.address,
		Handler:     router,
		ReadTimeout: readHeaderTimeout,
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

func (s *httpServer) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Read the file from the request body
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при чтении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, header.Filename)
	out, err := os.Create(tempFile)
	if err != nil {
		http.Error(w, "Ошибка при создании файла", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Copy data to the temp file
	if _, err = io.Copy(out, file); err != nil {
		http.Error(w, "Ошибка при сохранении файла", http.StatusInternalServerError)
		return
	}

	// Обработка файла
	recognizeResult, err := s.service.HandleFile(header, tempDir, tempFile)
	if err != nil {
		http.Error(w, "Ошибка при обработке файла", http.StatusInternalServerError)
		return
	}

	// Ответ от клиента
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(recognizeResult))

}
