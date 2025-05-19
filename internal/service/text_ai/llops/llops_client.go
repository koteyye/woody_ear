package llops

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLOpsClient клиент для работы с внешним LLMOps сервисом (аналогично Claude API)
type LLOpsClient struct {
	baseURL string
	token   string
	modelID string
	client  *http.Client
}

// Структура запроса к LLMOps API (формат Claude API)
type llmopsRequest struct {
	Model     string    `json:"model"`
	Messages  []message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

// Структура сообщения для запроса
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Структура ответа от LLMOps API (формат Claude API)
type llmopsResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewLLOpsClient создает новый клиент для работы с LLMOps
func NewLLOpsClient(baseURL, token, modelID string) (*LLOpsClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL не может быть пустым")
	}
	if token == "" {
		return nil, fmt.Errorf("token не может быть пустым")
	}
	if modelID == "" {
		return nil, fmt.Errorf("modelID не может быть пустым")
	}

	// Создаем HTTP клиент с таймаутом и отключенной проверкой сертификата
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   120 * time.Second, // Увеличиваем таймаут для больших моделей
		Transport: tr,
	}

	fmt.Printf("Создан LLMOps клиент для модели %s на сервере %s\n", modelID, baseURL)

	return &LLOpsClient{
		baseURL: baseURL,
		token:   token,
		modelID: modelID,
		client:  client,
	}, nil
}

// Analyze реализует интерфейс TextAI для работы с LLMOps
func (c *LLOpsClient) Analyze(ctx context.Context, text string) (string, error) {
	// Ограничиваем длину текста для анализа, если он слишком длинный
	maxTextLength := 10000 // Увеличиваем лимит для современных LLM
	truncatedText := text
	if len(text) > maxTextLength {
		truncatedText = text[:maxTextLength] + "..."
		fmt.Printf("Текст слишком длинный, обрезан до %d символов\n", maxTextLength)
	}

	// Создаем системный промпт и пользовательское сообщение
	systemPrompt := "Ты - помощник, который анализирует текст и делает его более читабельным и лаконичным. Твоя задача - проанализировать представленный текст, рассказать о его содержании и предложить улучшенную версию."
	userPrompt := "Проанализируй этот текст и сделай его более читабельным и лаконичным:\n\n" + truncatedText

	// Создаем запрос в формате Claude API
	reqBody := llmopsRequest{
		Model: c.modelID,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: 2000, // Ограничиваем длину ответа
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ошибка при сериализации запроса: %w", err)
	}

	// Формируем URL для запроса (используем формат Claude API)
	url := fmt.Sprintf("%s/v1/messages", c.baseURL)
	fmt.Printf("Отправляем запрос на %s\n", url)

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка при создании HTTP запроса: %w", err)
	}

	// Добавляем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.token) // Используем x-api-key для Claude API
	req.Header.Set("anthropic-version", "2023-06-01") // Версия API

	// Отправляем запрос
	fmt.Printf("Начинаем генерацию ответа через LLMOps, длина текста: %d символов\n", len(truncatedText))
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса к LLMOps: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLMOps вернул ошибку: статус %d, тело: %s", resp.StatusCode, string(bodyBytes))
	}

	// Читаем ответ
	var llmopsResp llmopsResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmopsResp); err != nil {
		return "", fmt.Errorf("ошибка при декодировании ответа: %w", err)
	}

	// Проверяем наличие ошибки в ответе
	if llmopsResp.Error.Message != "" {
		return "", fmt.Errorf("LLMOps вернул ошибку: %s", llmopsResp.Error.Message)
	}

	// Извлекаем текст из ответа
	var responseText string
	for _, content := range llmopsResp.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}

	if responseText == "" {
		return "", fmt.Errorf("LLMOps вернул пустой ответ")
	}

	fmt.Printf("Успешно получен ответ от LLMOps длиной %d символов\n", len(responseText))
	return responseText, nil
}
