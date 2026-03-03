# 🔭 InfraWatch

Production-grade infrastructure monitoring system built with Go.

## Architecture

```
┌─────────────┐     HTTPS/mTLS      ┌──────────────────┐     WebSocket     ┌──────────────┐
│   Agent(s)  │ ──────────────────▶  │  Central Server  │ ◀──────────────── │   Frontend   │
│  (Go binary)│   JSON metrics/5s    │   (Go + SQLite)  │   live updates    │   (React)    │
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
- **SQLite Storage**: WAL mode, time-series indexed, zero-config database
- **37 Unit Tests**: Comprehensive test coverage across 6 packages

## Quick Start

### Prerequisites
- Go 1.24+
- Make
- OpenSSL (for cert generation)

### 1. Clone and build
```bash
git clone https://github.com/nikhilanandd/Infrawatch.git
cd Infrawatch
make deps
make build
```

### 2. Generate mTLS certificates
```bash
make certs
```

### 3. Run the server
```bash
make run-server
```

### 4. Run the agent (separate terminal)
```bash
make run-agent
```

### 5. Test the API
```bash
# Health check
curl -k https://localhost:8443/health

# Login
curl -k -X POST https://localhost:8443/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Query nodes (use token from login response)
curl -k https://localhost:8443/api/nodes \
  -H "Authorization: Bearer <TOKEN>"
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/api/login` | No | Get JWT token |
| POST | `/api/metrics` | mTLS | Agent metric ingestion |
| GET | `/api/nodes` | JWT | List monitored nodes |
| GET | `/api/nodes/:id/metrics` | JWT | Query node metrics |
| GET | `/api/alerts` | JWT | Alert history |
| GET | `/ws` | JWT | WebSocket live updates |

## Testing

```bash
# Run all tests
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
│       ├── api/               # REST handlers + router
│       ├── auth/              # JWT + RBAC middleware
│       ├── store/             # SQLite storage
│       └── websocket/         # Live update hub
├── pkg/
│   ├── logger/                # Structured slog logger
│   └── models/                # Shared data models
├── configs/                   # YAML config files
├── scripts/                   # Certificate generation
├── deployments/               # Systemd units
├── web/                       # React frontend (TODO)
├── Dockerfile                 # Server container
├── Dockerfile.agent           # Agent container
├── docker-compose.yml         # Full stack
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
| Unit Tests | ✅ 37 passing |
| React Frontend | 🚧 In Progress |
| CI/CD | 🚧 Planned |

## License

MIT
