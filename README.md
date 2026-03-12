# 🔭 InfraWatch

[![CI](https://github.com/nikhilanandd/Infrawatch/actions/workflows/ci.yml/badge.svg)](https://github.com/nikhilanandd/Infrawatch/actions/workflows/ci.yml)

Production-grade infrastructure monitoring system built with Go.

## Architecture

```
┌─────────────┐     HTTPS/mTLS      ┌──────────────────┐     WebSocket     ┌──────────────┐
│   Agent(s)  │ ──────────────────▶  │  Central Server  │ ◀──────────────── │   Frontend   │
│  (Go binary)│   JSON metrics/5s    │   (Go + SQLite)  │   live updates    │   (React 18) │
└─────────────┘                      └──────────────────┘                   └──────────────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Alert Engine│
                                     │  Email/Slack │
                                     └─────────────┘
```

## Features

- **7 Metric Collectors**: CPU, Memory, Disk, Network, Docker, Systemd, Uptime
- **Concurrent Collection**: Goroutine-based parallel metric gathering
- **mTLS Security**: Mutual TLS authentication between agent and server
- **JWT + RBAC**: Role-based access control (admin/viewer)
- **Alert Engine**: Configurable rules (CPU > 80%, Disk > 85%, Service down)
- **Notifications**: Email (SMTP) and Slack webhook
- **Real-time**: WebSocket live metric streaming
- **React Dashboard**: Dark-themed SPA with Chart.js visualizations, served by the Go server
- **SQLite Storage**: WAL mode, time-series indexed, zero-config database
- **49 Unit & Integration Tests**: Comprehensive coverage across 7 packages
- **CI/CD**: GitHub Actions pipeline with Go + Node build verification

## Quick Start

### Prerequisites
- Go 1.24+
- Node.js 20+ (for frontend build)
- Make
- OpenSSL (for cert generation)

### 1. Clone and build
```bash
git clone https://github.com/nikhilanandd/Infrawatch.git
cd Infrawatch
make deps
make build
make build-web
```

### 2. Generate mTLS certificates
```bash
make certs
```

### 3. Run the server
```bash
make run-server
```
Open http://localhost:8443 — the Go server serves the React dashboard.

### 4. Run the agent (separate terminal)
```bash
make run-agent
```

### 5. Test the API
```bash
# Health check
curl -k https://localhost:8443/health

# Login (default: admin/admin)
curl -k -X POST https://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Query nodes (use token from login response)
curl -k https://localhost:8443/api/v1/nodes \
  -H "Authorization: Bearer <TOKEN>"
```

## Docker

```bash
# Build and start the full stack
docker compose up --build -d

# Check service health
docker compose ps

# View logs
docker compose logs -f server
docker compose logs -f agent

# Stop everything
docker compose down
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/api/v1/auth/login` | No | Authenticate, get JWT token |
| POST | `/api/v1/metrics` | mTLS | Agent metric ingestion |
| GET | `/api/v1/nodes` | JWT | List monitored nodes |
| GET | `/api/v1/metrics?node_id=X&from=T&to=T&limit=N` | JWT | Query node metrics |
| GET | `/api/v1/alerts?node_id=X&limit=N` | JWT | Alert history |
| GET | `/api/v1/alert-rules` | JWT | List alert rules |
| GET | `/api/v1/ws` | WS | WebSocket live updates |
| GET | `/` | No | React dashboard (SPA) |

## Testing

```bash
# Run all tests (unit + integration)
make test

# Verbose output
go test -v ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Project Structure

```
├── cmd/
│   ├── agent/main.go          # Agent entrypoint
│   └── server/main.go         # Server entrypoint
├── internal/
│   ├── agent/
│   │   ├── collector/         # 7 metric collectors
│   │   └── transport/         # mTLS HTTPS sender
│   ├── config/                # YAML config loader
│   └── server/
│       ├── alert/             # Rule engine + notifiers
│       ├── api/               # REST handlers + router + SPA serving
│       ├── auth/              # JWT + RBAC middleware
│       ├── store/             # SQLite storage
│       └── websocket/         # Live update hub
├── pkg/
│   ├── logger/                # Structured slog logger
│   └── models/                # Shared data models
├── configs/                   # YAML config files
├── scripts/                   # Certificate generation
├── deployments/               # Systemd units
├── web/                       # React 18 frontend (Chart.js, dark theme)
├── .github/workflows/         # CI/CD pipeline
├── Dockerfile                 # Server container (multi-stage)
├── Dockerfile.agent           # Agent container
├── docker-compose.yml         # Full stack orchestration
└── Makefile                   # Build targets
```

## Configuration

See `configs/server.yaml` and `configs/agent.yaml` for full configuration reference.

## Status

| Component | Status |
|-----------|--------|
| Go Agent | ✅ Complete |
| Go Server | ✅ Complete |
| SQLite DB | ✅ Complete |
| JWT Auth | ✅ Complete |
| Alert Engine | ✅ Complete |
| WebSocket | ✅ Complete |
| React Frontend | ✅ Complete |
| Unit Tests | ✅ 37 passing |
| Integration Tests | ✅ 12 passing |
| CI/CD | ✅ GitHub Actions |
| Docker | ✅ Multi-stage builds |

## License

This project is licensed under the [GNU General Public License v3.0](LICENSE).
