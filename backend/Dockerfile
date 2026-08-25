# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 先複製 go.mod / go.sum 利用 layer cache
COPY go.mod go.sum ./
RUN go mod download

# 複製所有原始碼並編譯
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/v1-p2p-exchange-service ./cmd/v1-p2p-exchange-service

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.20

# 時區與 TLS 憑證
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && \
    echo "Asia/Taipei" > /etc/timezone

WORKDIR /app

COPY --from=builder /app/bin/v1-p2p-exchange-service .
# config 由 K8s ConfigMap / Secret 掛載，這裡只放範本；相對路徑需與
# cmd/v1-p2p-exchange-service/main.go 的 -f 預設值（internal/constants/config.yaml）對齊
COPY cmd/v1-p2p-exchange-service/internal/constants/config.yaml internal/constants/config.yaml

# HTTP（REST API）與 WebSocket 各自獨立埠
EXPOSE 8888 8889

ENTRYPOINT ["/app/v1-p2p-exchange-service", "-f", "internal/constants/config.yaml"]
