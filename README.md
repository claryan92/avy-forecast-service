# 🏔️ Avalanche Forecast Service

A Go-based microservice that fetches, processes, and serves avalanche forecast data from multiple avalanche centers.  

---

## 📁 Project Structure

```
avy-forecast-service/
├── cmd/
│   └── server/                 # Main API entrypoint
├── internal/
│   ├── clients/                # HTTP clients for external APIs
│   ├── db/                     # Database layer (PostgreSQL)
│   ├── handlers/               # HTTP route handlers
│   ├── models/                 # Data structures and JSON models
│   ├── services/               # Core business logic (ForecastService)
│   └── utils/                  # Shared helpers/utilities
├── migrations/                 # SQL migrations (Flyway-compatible)
├── Dockerfile
├── Dockerfile.dev
├── docker-compose.yml
├── Makefile
├── .env
├── go.mod / go.sum
└── README.md
```

---

## ⚙️ Features

- Fetches forecast data concurrently from multiple avalanche centers.

---

## 🚀 Quick Start

### 1. Clone the repository
```bash
git clone https://github.com/yourusername/avy-forecast-service.git
cd avy-forecast-service
```

### 2. Set up environment variables
Create a `.env` file (you can use the template below):

```bash
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=avalanche
POSTGRES_HOST=db
POSTGRES_PORT=5432

API_PORT=8080
GO_ENV=development
```

### 3. Start the app
Run everything (API, DB, Flyway) in Docker:

```bash
make up
```

To rebuild cleanly:
```bash
make down
make up
```

To stop:
```bash
make down
```

---

## 🧪 Running Tests

You can run tests inside the dev container or locally.

**Using Docker (recommended):**
```bash
make test
```

**Locally:**
```bash
go test ./... -v
```

For coverage:
```bash
go test ./... -cover
```

---

## 🧱 API Overview

| Method | Endpoint             | Description                              |
|--------|----------------------|------------------------------------------|
| `GET`  | `/api/forecasts`     | Retrieve latest forecasts by zone/center |
| `GET`  | `/api/health`        | Health check endpoint                    |

Example response:
```json
[
  {
    "zone_id": "kootenai",
    "zone_name": "East Cabinet Mountains",
    "center": "Idaho Panhandle Avalanche Center",
    "issued_time": "2025-11-05T15:00:00Z",
    "today_danger": {
      "upper": 3,
      "middle": 2,
      "lower": 2,
      "valid_day": "current"
    }
  }
]
```

---

## 🧩 Makefile Commands

| Command | Description |
|----------|-------------|
| `make up` | Build and start all containers |
| `make down` | Stop and remove containers |
| `make dev` | Start the dev container with live reload |
| `make test` | Run all Go unit tests |
| `make migrate` | Create a new Flyway migration file |
| `make docs` | Generate `doc.go` files for all packages |

---

