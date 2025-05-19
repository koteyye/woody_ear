package speachRecognizer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	vosk "github.com/alphacep/vosk-api/go"
)

// Константы для работы с WAV файлами
const (
	defaultSampleRate = 16000 // Стандартная частота дискретизации для Vosk
	wavHeaderSize     = 44    // Размер заголовка WAV файла
)

// VoskService предоставляет сервис для распознавания речи с использованием Vosk API
type VoskService struct {
	model *vosk.VoskModel
}

// NewVoskService создает новый экземпляр сервиса распознавания речи
func NewVoskService() (*VoskService, error) {
	modelPath := os.Getenv("MODEL_PATH")
	if modelPath == "" {
		return nil, errors.New("MODEL_PATH не задан в переменных окружения")
	}

	// Проверяем существование модели
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("модель не найдена по пути %s: %w", modelPath, err)
	}

	model, err := vosk.NewModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать модель: %w", err)
	}

	return &VoskService{model: model}, nil
}

// RecognizeAudio распознает речь из WAV файла
func (vs *VoskService) RecognizeAudio(filePath string) (string, error) {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	// Определяем частоту дискретизации из WAV файла
	sampleRate, err := getSampleRateFromWAV(file)
	if err != nil {
		// Если не удалось определить частоту, используем стандартную
		sampleRate = defaultSampleRate
	}

	// Перемещаем указатель в начало файла
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("ошибка при перемещении указателя файла: %w", err)
	}

	// Читаем данные файла
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	// Создаем распознаватель с определенной частотой дискретизации
	recognizer, err := vosk.NewRecognizer(vs.model, float64(sampleRate))
	if recognizer == nil {
		return "", fmt.Errorf("не удалось создать распознаватель: %w", err)
	}
	defer recognizer.Free()

	// Устанавливаем параметры распознавания
	recognizer.SetMaxAlternatives(1)
	recognizer.SetWords(1) // 1 = true, 0 = false

	// Обрабатываем аудио данные
	acceptResult := recognizer.AcceptWaveform(data)
	if acceptResult != 1 {
		// Если AcceptWaveform не вернул 1, это может означать, что данные еще обрабатываются
		// или возникла ошибка. Мы все равно пытаемся получить результат.
	}

	// Получаем финальный результат
	finalResult := recognizer.FinalResult()
	if len(finalResult) == 0 {
		return "", errors.New("результат распознавания пуст")
	}

	return string(finalResult), nil
}

// getSampleRateFromWAV извлекает частоту дискретизации из WAV файла
func getSampleRateFromWAV(file *os.File) (int, error) {
	// Перемещаемся к позиции, где хранится частота дискретизации (обычно 24 байта от начала файла)
	if _, err := file.Seek(24, io.SeekStart); err != nil {
		return 0, err
	}

	// Читаем 4 байта, содержащие частоту дискретизации
	var sampleRateBytes [4]byte
	if _, err := file.Read(sampleRateBytes[:]); err != nil {
		return 0, err
	}

	// Преобразуем байты в число (little-endian)
	sampleRate := int(binary.LittleEndian.Uint32(sampleRateBytes[:]))

	// Проверяем, что частота дискретизации имеет разумное значение
	if sampleRate < 8000 || sampleRate > 192000 {
		return 0, fmt.Errorf("некорректная частота дискретизации: %d", sampleRate)
	}

	return sampleRate, nil
}
