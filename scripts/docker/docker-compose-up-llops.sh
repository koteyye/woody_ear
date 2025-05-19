#!/bin/bash

# Скрипт для запуска приложения Woody Ear с внешним LLOps сервисом

set -e

# Переходим в директорию со скриптом
cd "$(dirname "$0")"

echo "🔨 Запуск приложения с внешним LLOps сервисом..."
echo "⚠️ Первый запуск может занять некоторое время, так как нужно скачать модель vosk-api"
docker-compose -f docker-compose-llops.yml up -d --build

echo "✅ Приложение запущено и доступно по адресу http://localhost:8080"
echo "   Для загрузки аудиофайла используйте:"
echo "   curl -X POST -F \"file=@/path/to/your/audio.mp3\" http://localhost:8080/upload"
echo "   Или откройте веб-интерфейс: http://localhost:8080/test"
echo ""
echo "📋 Логи приложения:"
docker-compose -f docker-compose-llops.yml logs -f
