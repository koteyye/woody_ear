package textai

import "context"

// TextAI интерфейс для работы с AI-технологиями для анализа текста
type TextAI interface {
	// Analyze анализирование текста через AI модель
	Analyze(context.Context, string) (string, error)
}
