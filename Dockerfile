FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ffmpeg ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1000
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
COPY no_pfp.jpg /defaults/no_pfp.jpg

RUN mkdir -p /app/uploads/photos /app/uploads/audios /app/uploads/files \
    && chown -R appuser:appgroup /app /defaults

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
