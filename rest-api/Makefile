include .env

MIGRATIONS_DIR=./migrations
DSN=$(DATABASE_DSN)

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DSN)" up 1

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DSN)" down 1

migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database "$(DSN)" force $(v)

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) $(name)

migrate-drop:
	migrate -path $(MIGRATIONS_DIR) -database "$(DSN)" drop -f

migrate-path:
	migrate -path $(MIGRATIONS_DIR) -database "$(DSN)" version

swag:
	swag init --generalInfo ./cmd/api/main.go --output docs

build:
	go build -o api ./cmd/api

build-run: build
	godotenv ./api
