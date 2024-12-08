# Шаг 1: Используем официальный образ Go версии 1.21 для сборки приложения
FROM golang:1.22-alpine AS build

# Устанавливаем необходимые пакеты для работы с CGO и SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum, чтобы закешировать зависимости на этапе сборки
COPY go.mod go.sum ./

# Загружаем все зависимости
RUN go mod download

# Копируем весь исходный код проекта в контейнер
COPY . .

# Сборка приложения
RUN go build -o /app/url-shortener ./cmd/url-shortener

# Шаг 2: Создаем минимальный образ для запуска приложения
FROM alpine:latest

# Устанавливаем временную зону и настройки локали (опционально)
RUN apk add --no-cache tzdata

RUN mkdir -p /usr/local/bin/storage/db

# Устанавливаем переменные окружения
ENV APP_ENV=production
ENV CONFIG_PATH=/usr/local/bin/config/local.yaml
ENV APP_PORT=8080

# Копируем бинарник приложения из предыдущего шага
COPY --from=build /app/url-shortener /usr/local/bin/url-shortener

# Копируем конфигурационный файл в рабочую директорию
COPY --from=build /app/config/local.yaml /usr/local/bin/config/local.yaml

# Указываем рабочую директорию
WORKDIR /usr/local/bin

# Указываем команду для запуска приложения
CMD ["url-shortener"]

# Открываем порт (если приложение использует сетевой порт)
EXPOSE 8080