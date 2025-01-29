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
ENV VOSK_PATH=/app/model
ENV LD_LIBRARY_PATH=/app/model/lib
ENV CGO_CPPFLAGS="-I$VOSK_PATH/include -I$VOSK_PATH/pkg/vosk-api"
ENV CGO_LDFLAGS="-L$VOSK_PATH/lib -lvosk"

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
COPY --from=builder /app/model/lib /usr/lib
COPY --from=builder /app/model/include /usr/include
COPY --from=builder /app/pkg/vosk-api /usr/include/vosk-api

# Задать переменные окружения в конечном образе
ENV VOSK_PATH=/app/model
ENV LD_LIBRARY_PATH=$VOSK_PATH/lib:$LD_LIBRARY_PATH
ENV CGO_CPPFLAGS="-I$VOSK_PATH/include -I$VOSK_PATH/pkg/vosk-api"
ENV CGO_LDFLAGS="-L$VOSK_PATH/lib -lvosk"

# Запуск приложения
CMD ["./app"]