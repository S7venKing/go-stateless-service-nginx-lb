# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o status-service

# Runtime stage
FROM scratch

COPY --from=builder /app/status-service /status-service

EXPOSE 3000

ENTRYPOINT ["/status-service"]