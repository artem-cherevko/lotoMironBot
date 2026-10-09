# ===== Build =====
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o lotoMironBot ./cmd/app/

# ===== Runtime =====
FROM alpine:3.22

WORKDIR /app

# HTTPS certificates for Telegram API
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/lotoMironBot .

ENTRYPOINT ["./lotoMironBot"]
