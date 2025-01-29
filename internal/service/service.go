package service

import (
	"fmt"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	speachRecognizer "woody_ear/internal/service/speach_recognizer"
)

type Service struct {
	recognizer *speachRecognizer.VoskService
}

func NewService() (*Service, error) {
	recognizer, err := speachRecognizer.NewVoskService()
	if err != nil {
		return nil, fmt.Errorf("can't create recognizer: %w", err)
	}
	return &Service{recognizer: recognizer}, nil
}

func (s *Service) HandleFile(header *multipart.FileHeader, mp3FilePath, mp3file string) (string, error) {
	var readyFilePath string
	fileExtension := strings.ToLower(filepath.Ext(header.Filename))

	switch fileExtension {
	case ".mp3":
		wavFilePath, err := convertMP3ToWav(header, fileExtension, mp3FilePath, mp3file)
		if err != nil {
			return "", fmt.Errorf("can't convert mp3 to wav: %w", err)
		}
		readyFilePath = wavFilePath
	case ".wav":
		readyFilePath = filepath.Join(mp3FilePath, header.Filename)
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileExtension)
	}

	recognizeResult, err := s.recognizer.RecognizeAudio(readyFilePath)
	if err != nil {
		return "", fmt.Errorf("can't recognize audio: %w", err)
	}

	err = os.Remove(readyFilePath)
	if err != nil {
		return "", fmt.Errorf("can't remove temporary file: %w", err)
	}

	return recognizeResult, nil
}

func convertMP3ToWav(header *multipart.FileHeader, fileExtension, mp3FilePath, mp3file string) (wavFilePath string, err error) {
	wavFilePath = filepath.Join(mp3FilePath, strings.TrimSuffix(header.Filename, fileExtension)+".wav")

	cmd := exec.Command("sox", mp3file, wavFilePath)
	errOutput, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error converting mp3 to wav: %w, %s", err, string(errOutput))
	}

	return wavFilePath, nil
}
