# ── Stage 1: Build ────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 의존성 먼저 캐시
COPY go.mod ./
RUN go mod tidy

# 소스 복사 후 빌드
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o todoapi .

# ── Stage 2: Runtime ──────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache tzdata ca-certificates
ENV TZ=Asia/Seoul

WORKDIR /app
COPY --from=builder /app/todoapi .

EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./todoapi"]
