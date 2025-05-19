package gollama

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-skynet/go-llama.cpp"
)

type LocalLLMClient struct {
	model    *llama.LLama
	modelPath string
}

func NewLocalLLMClient(modelPath string) (*LocalLLMClient, error) {
	// Проверяем существование файла модели
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл модели не найден по пути: %s", modelPath)
	}
	
	// Выводим информацию о загружаемой модели
	fmt.Printf("Загружаем LLM модель из: %s\n", modelPath)
	
	// Используем базовую инициализацию модели
	model, err := llama.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("не могу загрузить модель: %w", err)
	}

	return &LocalLLMClient{
		model:     model,
		modelPath: modelPath,
	}, nil
}

func (c *LocalLLMClient) Analyze(ctx context.Context, text string) (string, error) {
	// Ограничиваем длину текста для анализа, если он слишком длинный
	maxTextLength := 1000
	truncatedText := text
	if len(text) > maxTextLength {
		truncatedText = text[:maxTextLength] + "..."
		fmt.Printf("Текст слишком длинный, обрезан до %d символов\n", maxTextLength)
	}
	
	prompt := "Проанализируй представленный текст и расскажи о нем, а также сделай его вариант более читабельным и лаконичным: " + truncatedText
	
	// Проверяем, что модель была успешно загружена
	if c.model == nil {
		return "", fmt.Errorf("модель не была загружена, проверьте путь к файлу модели: %s", c.modelPath)
	}
	
	fmt.Printf("Начинаем генерацию ответа с помощью LLM, длина текста: %d символов\n", len(truncatedText))
	
	// Настраиваем параметры генерации для более стабильной работы
	// Уменьшаем количество токенов и потоков для снижения нагрузки на память
	response, err := c.model.Predict(
		prompt, 
		llama.SetTemperature(0.5),     // Снижаем температуру для более предсказуемых результатов
		llama.SetTokens(256),          // Уменьшаем количество токенов еще сильнее
		llama.SetThreads(1),           // Используем только один поток
		llama.SetTopK(40),             // Ограничиваем выбор токенов
		llama.SetTopP(0.95),           // Устанавливаем вероятностный порог
	)
	if err != nil {
		fmt.Printf("Ошибка при генерации ответа: %v\n", err)
		return "", fmt.Errorf("ошибка при генерации ответа: %w", err)
	}
	
	fmt.Printf("Успешно сгенерирован ответ длиной %d символов\n", len(response))
	
	// Убираем промпт из ответа, если он там есть
	cleanResponse := strings.TrimPrefix(response, prompt)
	return strings.TrimSpace(cleanResponse), nil
}

func (c *LocalLLMClient) Close() {
	if c.model != nil {
		c.model.Free()
	}
}
