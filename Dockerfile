# Используем официальный образ Golang для сборки
FROM golang:latest AS builder

# Установить необходимые пакеты и зависимости
RUN apt-get update && apt-get install -y \
    unzip \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Установим директорию работы
WORKDIR /app

# Копируем все файлы вашего проекта в контейнер
COPY . .

# Скачиваем модели Vosk и распаковываем их
RUN wget https://alphacephei.com/vosk/models/vosk-model-small-ru-0.22.zip \
    && unzip vosk-model-small-ru-0.22.zip -d model

# Установить переменные окружения для пути к библиотекам Vosk
ENV VOSK_PATH=/app/src
ENV LD_LIBRARY_PATH=/app/src
ENV CGO_CPPFLAGS="-I$VOSK_PATH"
ENV CGO_LDFLAGS="-L$VOSK_PATH -lvosk"

# Соберите приложение
RUN go build -o app ./cmd/woody_ear

# Используем более легковесный образ для запуска
FROM debian:buster-slim

# Установить необходимые зависимости, включая sox
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sox \
    && rm -rf /var/lib/apt/lists/*

# Установить рабочую директорию
WORKDIR /app

# Копируем скомпилированное приложение из стадии сборки
COPY --from=builder /app/app /app
COPY --from=builder /app/model /app/model

# Копируем библиотеки vosk
COPY --from=builder /app/vosk-linux /app

# Задать переменные окружения в конечном образе
ENV VOSK_PATH=/app
ENV LD_LIBRARY_PATH=$VOSK_PATH:$LD_LIBRARY_PATH
ENV CGO_CPPFLAGS="-I$VOSK_PATH"
ENV CGO_LDFLAGS="-L$VOSK_PATH"

# Запуск приложения
CMD ["./app"]