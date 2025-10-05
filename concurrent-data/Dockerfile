FROM golang:1.24-alpine
WORKDIR /app

RUN apk add --no-cache git ca-certificates bash
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN swag init --generalInfo ./cmd/api/main.go --output ./docs
RUN CGO_ENABLED=0 GOOS=linux go build -o processor ./cmd/api

EXPOSE 8090
CMD ["./processor"]
