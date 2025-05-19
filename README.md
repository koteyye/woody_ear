# Woody Ear

Woody Ear - приложение для распознавания аудио-файлов и их анализа с помощью искусственного интеллекта.

<div align="center">
  <img src="https://images.deepai.org/art-image/428ac88e608b4059a99a47e6680b95a6/create-a-minimalist-and-stylish-logo-style-illustrati.jpg" alt="Woody Ear Logo" style="border-radius: 15px; max-width: 300px;" />
</div>

## Возможности

- Поддержка форматов `.wav` и `.mp3` (MP3 автоматически конвертируется в WAV)
- Распознавание речи с использованием библиотеки Vosk API
- Анализ распознанного текста с помощью LLM (Llama 2)
- Поддержка двух режимов работы:
  - Локальная модель LLM (go-llama.cpp)
  - Внешний LLOps сервис (Ollama)
- Простой HTTP API для загрузки и обработки файлов

## Технологии

- Go 1.23+
- Vosk API для распознавания речи
- Llama 2 для анализа текста
- Docker для контейнеризации

## Запуск в Docker

Самый простой способ запустить приложение - использовать Docker:

> **⚠️ Важное примечание:** В текущей версии библиотека go-llama.cpp не всегда стабильна и может вызывать ошибки "inference failed" при обработке текста. Для продакшена и стабильной работы **настоятельно рекомендуется** использовать вариант с внешним LLOps сервисом (Ollama).

### Docker Compose (рекомендуется)

#### С локальной LLM моделью

```bash
# Запуск с локальной LLM моделью
./scripts/docker/docker-compose-up-local.sh
```

#### С внешним LLOps сервисом (Ollama) - РЕКОМЕНДУЕТСЯ

```bash
# Запуск с внешним LLOps сервисом (рекомендуемый вариант)
./scripts/docker/docker-compose-up-llops.sh
```

> 💡 **Рекомендация:** Этот вариант предпочтительнее для продакшена, так как:
> - Ollama более стабилен, чем go-llama.cpp
> - Автоматически управляет памятью
> - Может быть легко масштабирован на отдельный сервер с GPU
> - Образ Docker значительно меньше (~300MB вместо ~4.5GB)
> - Сборка образа происходит намного быстрее (не нужно скачивать модель и компилировать go-llama.cpp)
> - Решает проблему "inference failed", которая может возникать при использовании go-llama.cpp

#### Стандартный запуск (устаревший)

```bash
# Запуск с помощью стандартного Docker Compose
./scripts/docker/docker-compose-up.sh

# Или вручную
cd scripts/docker
docker-compose up -d
```

### Обычный Docker

```bash
# Запуск с помощью скрипта
./scripts/docker/build-and-run.sh

# Или вручную
docker build -t woody-ear -f scripts/docker/Dockerfile .
docker run -p 8080:8080 woody-ear
```

После запуска приложение будет доступно по адресу http://localhost:8080

## Локальная сборка и запуск

### Требования

- Go 1.23 или выше
- Vosk API и модель для русского языка
- Модель Llama 2 (формат GGUF)

### MacOS

1. Установите зависимости:

```bash
brew install go cmake
```

2. Скачайте модель Vosk для русского языка:

```bash
mkdir -p models/vosk-model
wget https://alphacephei.com/vosk/models/vosk-model-small-ru-0.22.zip
unzip vosk-model-small-ru-0.22.zip -d models/
mv models/vosk-model-small-ru-0.22/* models/vosk-model/
```

3. Скачайте модель Llama 2:

```bash
mkdir -p models
wget https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF/resolve/main/llama-2-7b-chat.Q4_K_M.gguf -O models/llama-2-7b-chat.gguf
```

4. Установите переменные окружения:

```bash
export MODEL_PATH="$(pwd)/models/vosk-model"
export LLM_MODEL_PATH="$(pwd)/models/llama-2-7b-chat.gguf"
export SERVER_ADDRESS="localhost:8080"
export DYLD_LIBRARY_PATH="$(pwd)/models/vosk-model/lib"
export CGO_CPPFLAGS="-I$(pwd)/models/vosk-model/include"
export CGO_LDFLAGS="-L$(pwd)/models/vosk-model/lib -lvosk"
```

5. Соберите и запустите приложение:

```bash
# Запуск с помощью скрипта
./scripts/run-local.sh

# Или вручную
go build -o woody_ear ./cmd/woody_ear
./woody_ear
```

### Linux

Процесс аналогичен MacOS, но используйте `LD_LIBRARY_PATH` вместо `DYLD_LIBRARY_PATH`.

## Использование

### Веб-интерфейс

После запуска приложения вы можете использовать веб-интерфейс для загрузки файлов:

```
http://localhost:8080/test
```

Этот интерфейс позволяет выбрать MP3 или WAV файл и отправить его на обработку.

### API

Вы также можете использовать API напрямую:

#### Загрузка аудиофайла

```bash
curl -X POST -F "file=@/path/to/your/audio.mp3" http://localhost:8080/upload
```

Ответ будет содержать распознанный текст и его анализ:

```json
{
  "reconize": "распознанный текст из аудиофайла",
  "aiOutputText": "анализ текста от LLM"
}
```

### Проверка статуса сервера

```bash
curl http://localhost:8080/
```

Должен вернуть сообщение "Woody Ear Service is running".

## Переменные окружения

### Общие параметры
- `MODEL_PATH`: Путь к модели Vosk
- `SERVER_ADDRESS`: Адрес и порт сервера (по умолчанию "localhost:8080")
- `DEBUG`: Установите "true" для включения отладочного режима
- `TEMP_DIR`: Директория для временных файлов

### Параметры AI
- `AI_TYPE`: Тип AI сервиса ("local" для локальной модели, "llops" для внешнего сервиса)

#### Для локальной модели (AI_TYPE=local)
- `LLM_MODEL_PATH`: Путь к модели Llama 2

#### Для внешнего LLOps сервиса (AI_TYPE=llops)
- `LLOPS_BASE_URL`: URL LLOps сервера (по умолчанию "http://llops-server:11434")
- `LLOPS_TOKEN`: Токен авторизации для LLOps сервера (если требуется)
- `LLOPS_MODEL_ID`: ID модели на LLOps сервере (по умолчанию "llama2")

## Структура проекта

- `cmd/woody_ear`: Точка входа приложения
- `internal/api/http`: HTTP API сервер
- `internal/service/converter`: Конвертация MP3 в WAV
- `internal/service/speach_recognizer`: Распознавание речи с Vosk
- `internal/service/text_ai`: Анализ текста с LLM
  - `internal/service/text_ai/gollama`: Реализация анализа текста с локальной LLM моделью
  - `internal/service/text_ai/llops`: Реализация анализа текста через внешний LLOps сервис
- `scripts`: Вспомогательные скрипты
  - `scripts/docker`: Файлы для Docker
    - `scripts/docker/Dockerfile`: Устаревший Dockerfile (для совместимости)
    - `scripts/docker/Dockerfile.local`: Dockerfile для сборки образа с локальной LLM моделью
    - `scripts/docker/Dockerfile.llops`: Dockerfile для сборки образа с внешним LLOps сервисом (легче и быстрее)
    - `scripts/docker/docker-compose.yml`: Стандартная конфигурация Docker Compose
    - `scripts/docker/docker-compose-local.yml`: Конфигурация для запуска с локальной LLM моделью
    - `scripts/docker/docker-compose-llops.yml`: Конфигурация для запуска с внешним LLOps сервисом
    - `scripts/docker/build-and-run.sh`: Скрипт для запуска с помощью Docker
    - `scripts/docker/docker-compose-up.sh`: Стандартный скрипт для запуска с помощью Docker Compose
    - `scripts/docker/docker-compose-up-local.sh`: Скрипт для запуска с локальной LLM моделью
    - `scripts/docker/docker-compose-up-llops.sh`: Скрипт для запуска с внешним LLOps сервисом
  - `scripts/run-local.sh`: Скрипт для локального запуска
  - `scripts/test-upload.html`: Веб-интерфейс для загрузки файлов
