# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 先複製 go.mod / go.sum 利用 layer cache
COPY go.mod go.sum ./
RUN go mod download

# 複製所有原始碼並編譯
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/p2p p2p.go

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.20

# 時區與 TLS 憑證
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && \
    echo "Asia/Taipei" > /etc/timezone

WORKDIR /app

COPY --from=builder /app/bin/p2p .
# config 由 K8s ConfigMap / Secret 掛載，這裡只放範本
COPY etc/config.yaml etc/config.yaml

EXPOSE 8888

ENTRYPOINT ["/app/p2p", "-f", "etc/config.yaml"]
