package converter

import (
	"fmt"
	"github.com/hajimehoshi/go-mp3"
	"github.com/youpy/go-wav"
	"io"
	"os"
	"runtime"
)

// Размер буфера для чтения/записи
const bufferSize = 8192 * 4 // 32KB буфер для лучшей производительности

// ConvertMP3ToWavGo конвертирует MP3 файл в WAV формат
// mp3Path - путь к исходному MP3 файлу
// wavPath - путь, куда сохранить WAV файл
// Возвращает путь к созданному WAV файлу и ошибку, если она возникла
func ConvertMP3ToWavGo(mp3Path, wavPath string) (string, error) {
	// Открываем MP3 файл
	mp3File, err := os.Open(mp3Path)
	if err != nil {
		return "", fmt.Errorf("не удалось открыть MP3 файл: %w", err)
	}
	defer mp3File.Close()

	// Создаем декодер MP3
	decoder, err := mp3.NewDecoder(mp3File)
	if err != nil {
		return "", fmt.Errorf("не удалось создать MP3 декодер: %w", err)
	}

	// Создаем WAV файл
	wavFile, err := os.Create(wavPath)
	if err != nil {
		return "", fmt.Errorf("не удалось создать WAV файл: %w", err)
	}
	defer wavFile.Close()

	// Получаем параметры из MP3 файла
	sampleRate := decoder.SampleRate()
	numChannels := 2 // MP3 обычно стерео
	bitsPerSample := 16
	
	// Создаем WAV писатель
	writer := wav.NewWriter(wavFile, uint32(decoder.Length()), uint16(numChannels), uint32(sampleRate), uint16(bitsPerSample))
	
	// Создаем буфер для чтения/записи
	buf := make([]byte, bufferSize)
	
	// Конвертируем данные
	totalBytes := 0
	for {
		// Читаем данные из MP3
		n, err := decoder.Read(buf)
		if n > 0 {
			// Записываем данные в WAV
			if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
				return "", fmt.Errorf("ошибка при записи в WAV файл: %w", writeErr)
			}
			totalBytes += n
		}
		
		// Проверяем на конец файла
		if err == io.EOF {
			break
		}
		
		// Проверяем на другие ошибки
		if err != nil {
			return "", fmt.Errorf("ошибка при чтении MP3 файла: %w", err)
		}
		
		// Периодически запускаем сборщик мусора для больших файлов
		if totalBytes > 10*1024*1024 { // Каждые 10MB
			totalBytes = 0
			runtime.GC() // Подсказка сборщику мусора
		}
	}
	
	return wavPath, nil
}
