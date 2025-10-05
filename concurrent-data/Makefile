include .env
swag:
	swag init --generalInfo ./cmd/api/main.go --output docs

build:
	go build -o api ./cmd/api

build-run: build
	godotenv ./api

docker-build:
	docker build -t concurrent-data:latest .

docker-run:
	docker run --rm -p 8090:8090 \
		-v $(PWD)/sample:/app/sample \
		--env-file .env \
		concurrent-data:latest