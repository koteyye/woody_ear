package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const baseModel = "qwen2.5:3b"

type OllamaHTTPClient struct {
	client  *http.Client
	url     *url.URL
	aimodel string
}

func NewOllamaClient(configUrl string) (*OllamaHTTPClient, error) {
	u, err := url.Parse(configUrl)
	if err != nil {
		return nil, fmt.Errorf("can't parse url: %w", err)
	}

	return &OllamaHTTPClient{
		client: &http.Client{},
		url:    u,
		aimodel: baseModel,
	}, nil
}

func (c *OllamaHTTPClient) Analyze(ctx context.Context, text string) (string, error) {
	body, err := json.Marshal(request{
		Model:  c.aimodel,
		Prompt: "Проанализируй представленный текст и расскажи о нем, а также сделай его вариант более читабельным и лаконичным" + text,
	})
	if err != nil {
		return "", fmt.Errorf("can't marshal request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url.String(),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("can't create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("can't do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("can't decode response: %w", err)
	}

	return result.Response, nil
}
