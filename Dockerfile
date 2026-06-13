# Стадия сборки
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

# Финальный образ
FROM alpine:3.19
RUN apk add --no-cache ffmpeg ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1000
WORKDIR /app
COPY --from=builder /app/server .
# Положите файл no_pfp.jpg в папку backend рядом с Dockerfile
COPY no_pfp.jpg /defaults/no_pfp.jpg

# Создаём структуру томов
RUN mkdir -p /app/uploads/photos /app/uploads/audios /app/uploads/files \
    && chown -R appuser:appgroup /app /defaults

# При запуске: дефолтный файл, симлинки, запуск сервера
CMD sh -c "\
    if [ ! -f /app/uploads/photos/no_pfp.jpg ]; then \
        cp /defaults/no_pfp.jpg /app/uploads/photos/no_pfp.jpg; \
    fi && \
    rm -rf /app/photos /app/audios /app/files && \
    ln -sfn /app/uploads/photos /app/photos && \
    ln -sfn /app/uploads/audios /app/audios && \
    ln -sfn /app/uploads/files /app/files && \
    ./server"

USER appuser
EXPOSE 8080