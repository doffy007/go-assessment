include .env
swag:
	swag init --generalInfo ./cmd/api/main.go --output docs

build:
	go build -o api ./cmd/api

build-run: build
	godotenv ./api