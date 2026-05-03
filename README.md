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
| Storage | MongoDB (`mongo-driver/v2`) |
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
├── .github/workflows/    # GitHub Actions CI pipeline
├── cmd/                  # Application entry point
├── internal/             # Core business logic
│   ├── handler/          # HTTP handlers
│   ├── repository/       # DB & cache access layer
│   └── service/          # Business logic layer
├── docker-compose.yaml   # Full stack Docker setup
├── go.mod
└── go.sum
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
POSTGRES_USER=postgres
POSTGRES_PW=yourpassword
POSTGRES_DB=urlshortener
POSTGRES_PORT=5432

JWT_SECRET=your_jwt_secret
```

### 3. Start the stack

```bash
docker compose up -d
```

This starts PostgreSQL, Redis, and Redis Commander (UI at `http://localhost:8081`).

### 4. Run the application

```bash
go run cmd/main.go
```

---

## API Overview

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/shorten` | Create a short URL (auth required) |
| `GET` | `/:code` | Redirect to original URL |
| `DELETE` | `/:code` | Delete a short URL (auth required) |
| `POST` | `/login` | Obtain a JWT token |

---

## CI/CD

GitHub Actions pipeline runs on every push to `main`, automatically building and testing the service.

---

## Author

**Hady Gustianto** — [LinkedIn](https://linkedin.com/in/hadygustianto) · [GitHub](https://github.com/hadygust)
