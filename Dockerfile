# ===== Build =====
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Зависимости
COPY go.mod go.sum ./
RUN go mod download

# Исходники
COPY . .

# Сборка
RUN CGO_ENABLED=0 GOOS=linux go build -o lotoMironBot ./cmd/app/

# ===== Runtime =====
FROM alpine:latest

WORKDIR /app

# Сертификаты нужны для HTTPS/TG API
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/lotoMironBot .

ENTRYPOINT ["./lotoMironBot"]