package speachRecognizer

import (
	"errors"
	"fmt"
	"os"

	vosk "github.com/alphacep/vosk-api/go"
)

type VoskService struct {
	model *vosk.VoskModel
}

func NewVoskService() (*VoskService, error) {
	modelPath := os.Getenv("VOSK_PATH") + "/vosk-model"
	model, err := vosk.NewModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("can't create model: %w", err)
	}

	return &VoskService{model: model}, nil
}

func (vs *VoskService) RecognizeAudio(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("can't read file: %w", err)
	}

	recognizer, err := vosk.NewRecognizer(vs.model, 64000)
	if recognizer == nil {
		return "", fmt.Errorf("can't create recognizer: %w", err)
	}
	defer recognizer.Free()

	recognizer.AcceptWaveform(data)
	result := recognizer.FinalResult()
	if len(result) == 0 {
		return "", errors.New("result is empty")
	}

	return string(result), nil
}