FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo -ldflags "-s -w" -o /app/server ./cmd/api

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates wget \
	&& addgroup -S app \
	&& adduser -S app -G app

COPY --from=builder /app/server /app/server

EXPOSE 8080

USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget -qO- http://localhost:${PORT:-8080}/api/keep-alive || exit 1

CMD ["/app/server"]
