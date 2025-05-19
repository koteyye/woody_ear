#!/bin/bash

# Скрипт для сборки и запуска приложения Woody Ear в Docker

set -e

# Переходим в корневую директорию проекта
cd "$(dirname "$0")/../.."
ROOT_DIR=$(pwd)

echo "🔨 Сборка Docker образа..."
docker build -t woody-ear -f scripts/docker/Dockerfile .

echo "🚀 Запуск контейнера..."
docker run -p 8080:8080 woody-ear

echo "✅ Приложение запущено и доступно по адресу http://localhost:8080"
echo "   Для загрузки аудиофайла используйте:"
echo "   curl -X POST -F \"file=@/path/to/your/audio.mp3\" http://localhost:8080/upload"
echo "   Или откройте веб-интерфейс: http://localhost:8080/test"
