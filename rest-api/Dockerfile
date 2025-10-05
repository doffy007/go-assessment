# Stage 1: Builder
FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates bash
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN swag init --generalInfo ./cmd/api/main.go --output ./docs
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# Stage 2: Runtime
FROM alpine:3.18
WORKDIR /app

COPY --from=builder /app/api ./
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/cred ./cred

EXPOSE 8080
CMD ["./api"]
