# URL Shortener

A production-ready **URL shortening service** built with Go, featuring JWT authentication, Redis caching, PostgreSQL persistence, and a CI/CD pipeline via GitHub Actions. Supports QUIC/HTTP3 transport and is fully containerized with Docker.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP Framework | Gin |
| Router/Transport | QUIC (`quic-go`) |
| Auth | JWT (`golang-jwt/jwt`) |
| Cache | Redis 7 (`go-redis/v9`) |
| Database | PostgreSQL (`pgx/v5`, `sqlx`) |
| Config | `godotenv` |
| IDs | UUID (`google/uuid`) |
| CI/CD | GitHub Actions |
| Containerization | Docker Compose |

---

## Features

- **Shorten URLs** — generate unique short codes for any long URL
- **Redirect** — fast lookup and redirect via short code
- **Redis caching** — hot URLs served directly from cache, reducing DB load
- **JWT authentication** — protected endpoints for URL management
- **Persistent storage** — all URLs durably stored in PostgreSQL
- **UUID-based keys** — globally unique, collision-resistant short identifiers
- **CI/CD pipeline** — automated build and test via GitHub Actions
- **Docker Compose** — spin up the full stack (app + PostgreSQL + Redis + Redis Commander) with one command

---

## Project Structure

```
url-shortener/
├── .github/workflows/        # CI pipeline (GitHub Actions)
├── cmd/                      # Application entry point (main.go, app setup)
├── internal/                 # Core business logic
│   ├── auth/                 # Authentication module (handler, service, repo, middleware)
│   ├── url/                  # URL shortening core logic (handler, service, repo)
│   ├── redirect_log/         # Redirect tracking logic (service, repository)
│   ├── rate_limiter/         # Rate limiting middleware & logic
│   ├── cache/                # Redis cache abstraction
│   ├── dto/                  # Data Transfer Objects (request/response schemas)
│   ├── model/                # Database models / entities
│   ├── env/                  # Environment variable loading & config
│   └── migration/            # Database migrations (Goose + SQL files)
├── docker-compose.yaml       # Multi-service setup (Postgres, Redis, Redis Commander)
├── go.mod                   
├── go.sum                    
```

---

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose

### 1. Clone the repository

```bash
git clone https://github.com/hadygust/url-shortener.git
cd url-shortener
```

### 2. Configure environment

Create a `.env` file in the root:

```env
# PostgreSQL Configuration
POSTGRES_HOST=your_postgres_host
POSTGRES_PORT=5432
POSTGRES_USER=your_postgres_user
POSTGRES_PW=your_postgres_password
POSTGRES_DB=your_database_name

# JWT Configuration
JWT_SECRET=your_jwt_secret

# Goose Migration
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://user:password@host:port/dbname
GOOSE_MIGRATION_DIR=./internal/migration

# Redis Configuration
REDIS_ADDR=your_redis_host:6379
```

### 3. Start the stack

```bash
docker compose up -d
```

This starts PostgreSQL, Redis, and Redis Commander (UI at `http://localhost:8081`).

### 4. Run the application

```bash
go run ./cmd/*
```

---

## API Overview

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/url` | Create a short URL (auth required) |
| `GET` | `/:shortCode` | Redirect to original URL with given short code |
| `DELETE` | `/url/:shortCode` | Delete a short URL (auth required) |
| `POST` | `/auth/login` | Obtain a JWT token |

---

## CI/CD

GitHub Actions pipeline runs on every push to `main`, automatically building and testing the service.

---

## Author

**Hady Gustianto** — [LinkedIn](https://linkedin.com/in/hadygustianto) · [GitHub](https://github.com/hadygust)
