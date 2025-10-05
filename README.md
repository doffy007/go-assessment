# go-assessment

Take-home test for Go Programming Assessment.

---

## Project Overview
This project implements a Go-based system with the following components:

1. **REST API**: A simple blog system with user authentication, CRUD for posts and comments, input validation, and database integration with transactions.
2. **Concurrent Data Processor**: Worker pool-based processor for handling multiple CSV files simultaneously.

The project is fully containerized with Docker and includes database migrations, Swagger documentation, environment-based configuration, and basic test scaffolding.

---

## Table of Contents
1. [Setup & Running Instructions](#setup--running-instructions)
2. [Architecture Overview](#architecture-overview)
3. [Technology Choices](#technology-choices)
4. [Project Structure](#project-structure)
5. [Database & Migrations](#database--migrations)
6. [API Documentation](#api-documentation)
7. [Delivered vs Pending](#delivered-vs-pending)
8. [Known Limitations & Future Improvements](#known-limitations--future-improvements)
9. [Test Coverage](#test-coverage)

---

## Setup & Running Instructions

1. Clone the repository:
```bash
git clone <repo-url>
cd go-assessment


2. Create .env file in rest-api/ and concurrent-data/ and root directories with

# General Configuration
ENV_TYPE=dev
APP_NAME=test-api
APP_VERSION=1.0.0
PROJECT_NAME=test-api

# Server Configuration
SERVER_PORT=8080
REQUEST_TIMEOUT=20

# Database Configuration
DATABASE_URL=localhost
DATABASE_PORT=5432
DATABASE_NAME=rest_api_go
DATABASE_USERNAME=fauzan
DATABASE_PASSWORD=1234
DATABASE_SCHEMA=public
DATABASE_CONNECTION_TIMEOUT=20s
DATABASE_MAX_IDLE_CONNECTION=5
DATABASE_MAX_OPEN_CONNECTION=10
DATABASE_DEBUG_MODE=true
DATABASE_DSN=postgres://${DATABASE_USERNAME}:${DATABASE_PASSWORD}@db:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=disable
DATABASE_PING_INTERVAL=10

# Sonyflake
SONYFLAKE_IP=127.0.0.1

# Suspicious Email Detection
SUSPICIOUS_EMAIL_DETECTION_ENABLE=false

# JWT Keys
JWT_PUBLIC_KEY=./cred/public.pem
JWT_PRIVATE_KEY=./cred/private.pem

# Processor
CSV_PATH=/app/sample
WORKERS=4
PORT=8090

3. Build and start Docker containers:
docker compose up --build

4. Run database migrations:
docker compose run --rm migrate

5. Access REST API Swagger documentation:
http://localhost:8080/swagger/index.html

6. Access Concurrent Data Swagger documentation:
http://localhost:8090/swagger/index.html


Technology Choices

- Go: Performance, concurrency
- Gin: Lightweight HTTP web framework
- PostgreSQL: Relational database with ACID compliance
- Docker & Docker Compose: Simplified deployment
- Swagger/OpenAPI: API documentation
- Migrate: Database schema migrations
- Jwt: Validations auth

Project Structure

go-assessment/
├── rest-api/                     # REST API service
│   ├── cmd/api/main.go           # Entry point
│   ├── internal/                 # Application logic
│   │   ├── auth/                 # Authentication logic
│   │   ├── post/                 # Post CRUD operations
│   │   ├── comment/              # Comment CRUD operations
│   │   └── user/                 # User management
│   ├── migrations/               # SQL migration files
│   ├── docs/                     # Swagger documentation
│   ├── .env                      # Environment variables
│   ├── go.mod
│   └── Makefile
├── concurrent-data/              # Concurrent CSV processor
│   ├── cmd/processor/main.go
│   ├── internal/                 # Processor logic
│   ├── docs/                     # Optional processor docs
│   ├── sample/                   # Sample CSV files
│   ├── api                       # Built binary for testing
│   ├── Dockerfile
│   ├── Makefile
│   ├── go.mod
│   └── go.sum
├── docker-compose.yml
└── README.md
