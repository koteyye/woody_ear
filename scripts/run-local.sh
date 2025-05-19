#!/bin/bash

# Скрипт для локальной сборки и запуска приложения Woody Ear

set -e

# Переходим в корневую директорию проекта
cd "$(dirname "$0")/.."
ROOT_DIR=$(pwd)

# Проверяем наличие необходимых директорий
mkdir -p $ROOT_DIR/models/vosk-model
mkdir -p $ROOT_DIR/tmp

# Проверяем наличие модели Vosk
if [ ! -d "$ROOT_DIR/models/vosk-model/lib" ]; then
  echo "⬇️ Скачиваем модель Vosk для русского языка..."
  wget -q --show-progress https://alphacephei.com/vosk/models/vosk-model-small-ru-0.22.zip
  unzip -q vosk-model-small-ru-0.22.zip -d $ROOT_DIR/models/
  mv $ROOT_DIR/models/vosk-model-small-ru-0.22/* $ROOT_DIR/models/vosk-model/
  rm -rf $ROOT_DIR/models/vosk-model-small-ru-0.22 vosk-model-small-ru-0.22.zip
fi

# Проверяем наличие модели LLM
if [ ! -f "$ROOT_DIR/models/llama-2-7b-chat.gguf" ]; then
  echo "⬇️ Скачиваем модель Llama 2..."
  wget -q --show-progress https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF/resolve/main/llama-2-7b-chat.Q4_K_M.gguf -O $ROOT_DIR/models/llama-2-7b-chat.gguf
fi

# Копируем HTML файл для веб-интерфейса
cp $ROOT_DIR/scripts/test-upload.html $ROOT_DIR/test-upload.html

# Устанавливаем переменные окружения
export MODEL_PATH="$ROOT_DIR/models/vosk-model"
export LLM_MODEL_PATH="$ROOT_DIR/models/llama-2-7b-chat.gguf"
export SERVER_ADDRESS="localhost:8080"
export TEMP_DIR="$ROOT_DIR/tmp"

# Определяем переменные окружения в зависимости от ОС
if [[ "$OSTYPE" == "darwin"* ]]; then
  # MacOS
  export DYLD_LIBRARY_PATH="$MODEL_PATH/lib"
else
  # Linux
  export LD_LIBRARY_PATH="$MODEL_PATH/lib"
fi

export CGO_CPPFLAGS="-I$MODEL_PATH/include"
export CGO_LDFLAGS="-L$MODEL_PATH/lib -lvosk"

# Собираем приложение
echo "🔨 Сборка приложения..."
cd $ROOT_DIR
go build -o woody_ear ./cmd/woody_ear

# Запускаем приложение
echo "🚀 Запуск приложения..."
./woody_ear

echo "✅ Приложение запущено и доступно по адресу http://localhost:8080"
echo "   Для загрузки аудиофайла используйте:"
echo "   curl -X POST -F \"file=@/path/to/your/audio.mp3\" http://localhost:8080/upload"
echo "   Или откройте веб-интерфейс: http://localhost:8080/test"
