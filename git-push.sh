#!/bin/bash
set -e

REPO_URL="git@github.com:Nikhilanandd/Infrawatch.git"
BRANCH="dev"

echo "🚀 InfraWatch — Structured Git Push Script"
echo "============================================"

# Initialize repo if not already
if [ ! -d ".git" ]; then
    echo "📦 Initializing git repository..."
    git init
    git remote add origin "$REPO_URL"
else
    echo "📦 Git repo already initialized"
    # Ensure remote is set
    git remote set-url origin "$REPO_URL" 2>/dev/null || git remote add origin "$REPO_URL"
fi

# Create and switch to dev branch
git checkout -b "$BRANCH" 2>/dev/null || git checkout "$BRANCH"

echo ""
echo "=== Commit 1: Project scaffolding ==="
git add go.mod go.sum
git add Makefile
git add .gitignore 2>/dev/null || true
git commit -m "chore: initialize Go module and project scaffolding

- Set up go.mod with module github.com/nikhilanandd/infrawatch
- Add dependencies: jwt, websocket, sqlite3, bcrypt, yaml
- Add Makefile with build, run, test, docker targets" --allow-empty

echo ""
echo "=== Commit 2: Data models ==="
git add pkg/models/metrics.go
git add pkg/models/types.go
git commit -m "feat(models): add core data models for metrics pipeline

- MetricPayload with CPU, Memory, Disk, Network, Uptime stats
- ContainerStats for Docker monitoring
- ServiceStatus for systemd service tracking
- Full JSON serialization tags for API transport"

echo ""
echo "=== Commit 3: Structured logger ==="
git add pkg/logger/logger.go
git commit -m "feat(logger): add structured JSON logging with slog

- Configurable log levels (debug, info, warn, error)
- JSON output format for production use
- Package-level initialization function"

echo ""
echo "=== Commit 4: Configuration system ==="
git add internal/config/config.go
git add configs/ 2>/dev/null || true
git commit -m "feat(config): add YAML-based configuration for agent and server

- ServerConfig: listen addr, TLS paths, JWT secret, DB path, alerting
- AgentConfig: node ID, server URL, collection interval, services list
- Email and Slack alerting configuration support"

echo ""
echo "=== Commit 5: Agent metric collectors ==="
git add internal/agent/collector/cpu.go
git add internal/agent/collector/memory.go
git add internal/agent/collector/disk.go
git add internal/agent/collector/network.go
git add internal/agent/collector/docker.go
git add internal/agent/collector/systemd.go
git add internal/agent/collector/uptime.go
git commit -m "feat(agent): implement 7 concurrent metric collectors

- CPU: parse /proc/stat with per-core usage and load averages
- Memory: parse /proc/meminfo for RAM and swap stats
- Disk: syscall.Statfs for partition usage
- Network: parse /proc/net/dev for interface stats
- Docker: Unix socket API for container stats
- Systemd: systemctl show for service status
- Uptime: parse /proc/uptime
- All collectors designed for goroutine-based concurrent execution"

echo ""
echo "=== Commit 6: Agent transport layer ==="
git add internal/agent/transport/sender.go
git commit -m "feat(agent): add mTLS HTTPS transport for metric delivery

- Mutual TLS authentication with configurable cert/key/CA
- JSON payload serialization
- HTTP POST to central server /api/metrics endpoint
- Connection timeout and error handling"

echo ""
echo "=== Commit 7: Agent entrypoint ==="
git add cmd/agent/main.go
git commit -m "feat(agent): add main entrypoint with graceful shutdown

- YAML config loading via CLI flag
- Goroutine-based concurrent metric collection
- Configurable collection interval (default 5s)
- Signal handling for SIGINT/SIGTERM graceful shutdown
- Structured logging throughout"

echo ""
echo "=== Commit 8: SQLite storage layer ==="
git add internal/server/store/sqlite.go
git add internal/server/store/types.go
git commit -m "feat(server): implement SQLite storage with WAL mode

- Auto-migration: creates nodes, metrics, alerts, users, alert_rules tables
- Time-series indexed metrics storage with JSON payload
- CRUD operations for nodes, metrics, alerts, users
- Default admin user seeded on first run (admin/admin)
- WAL mode enabled for concurrent read performance
- Configurable data retention and query windowing"

echo ""
echo "=== Commit 9: JWT authentication and RBAC ==="
git add internal/server/auth/jwt.go
git add internal/server/auth/middleware.go
git commit -m "feat(server): add JWT authentication with role-based access control

- JWT token generation with username, role, and 24h expiry
- Token validation and claims extraction
- bcrypt password hashing and verification
- AuthMiddleware: validates Bearer token on protected routes
- RoleMiddleware: restricts endpoints by role (admin/viewer)
- Proper 401/403 error responses"

echo ""
echo "=== Commit 10: Alert engine and notifiers ==="
git add internal/server/alert/engine.go
git add internal/server/alert/notifiers.go
git commit -m "feat(server): implement alert rule engine with email and Slack notifiers

- Rule evaluation: CPU > 80%, Disk > 85%, Service down
- Alert deduplication and severity classification
- Email notifier via SMTP with configurable recipients
- Slack notifier via webhook URL with formatted messages
- Extensible rule and notifier architecture"

echo ""
echo "=== Commit 11: WebSocket hub for live updates ==="
git add internal/server/websocket/hub.go
git commit -m "feat(server): add WebSocket hub for real-time metric streaming

- Hub pattern with client registration/unregistration
- Broadcast metrics and alerts to all connected clients
- Thread-safe client management
- Client count tracking"

echo ""
echo "=== Commit 12: REST API handlers and router ==="
git add internal/server/api/handlers.go
git add internal/server/api/router.go
git commit -m "feat(server): implement REST API with full route configuration

Routes:
  POST /api/login          - JWT authentication
  POST /api/metrics        - Agent metric ingestion
  GET  /api/nodes          - List monitored nodes
  GET  /api/nodes/:id/metrics - Query node metrics with time window
  GET  /api/alerts         - Alert history
  GET  /health             - Health check endpoint
  GET  /ws                 - WebSocket upgrade for live updates

- CORS middleware for frontend integration
- JWT auth middleware on protected routes
- Request logging and error handling"

echo ""
echo "=== Commit 13: Server entrypoint ==="
git add cmd/server/main.go
git commit -m "feat(server): add main entrypoint with graceful shutdown

- YAML config loading
- SQLite database initialization
- TLS server with configurable cert/key
- Alert engine background goroutine
- WebSocket hub startup
- Signal handling for graceful shutdown with 10s timeout"

echo ""
echo "=== Commit 14: Unit tests (37 tests, 6 packages) ==="
git add pkg/models/metrics_test.go
git add pkg/logger/logger_test.go
git add internal/config/config_test.go
git add internal/server/auth/auth_test.go
git add internal/server/store/store_test.go
git add internal/server/websocket/hub_test.go
git add internal/server/alert/engine_test.go 2>/dev/null || true
git commit -m "test: add comprehensive unit tests across all packages

- pkg/models: JSON serialization roundtrips, omitempty behavior
- pkg/logger: log level initialization, JSON output
- internal/config: YAML loading, missing files, invalid YAML, defaults
- internal/server/auth: JWT generate/validate, wrong secrets, password hashing
- internal/server/store: full SQLite CRUD (nodes, metrics, alerts, users)
- internal/server/websocket: hub creation, broadcast safety

37 tests passing across 6 packages"

echo ""
echo "=== Commit 15: DevOps files ==="
git add Dockerfile 2>/dev/null || true
git add Dockerfile.agent 2>/dev/null || true
git add docker-compose.yml 2>/dev/null || true
git add scripts/ 2>/dev/null || true
git add deployments/ 2>/dev/null || true
# Only commit if there are staged changes
git diff --cached --quiet || git commit -m "ops: add Docker, docker-compose, systemd units, and cert generation

- Multi-stage Dockerfile for server (Go build + Alpine runtime)
- Dockerfile.agent for lightweight agent image
- docker-compose.yml with server + agent services
- systemd service units for production deployment
- mTLS certificate generation script (CA + server + agent certs)
- Makefile targets for build, test, docker, and cert generation"

echo ""
echo "=== Commit 16: Documentation ==="
git add README.md 2>/dev/null || true
git diff --cached --quiet || git commit -m "docs: add project README with architecture and setup instructions

- Project overview and architecture diagram
- Quick start guide (local and Docker)
- API endpoint documentation
- Configuration reference
- Development and testing instructions"

echo ""
echo "=== Commit 17: Any remaining files ==="
git add -A
git diff --cached --quiet || git commit -m "chore: add remaining project files and configurations"

echo ""
echo "============================================"
echo "📊 Commit Summary:"
git log --oneline
echo ""
echo "============================================"
echo "🚀 Pushing to origin/$BRANCH..."
git push -u origin "$BRANCH"

echo ""
echo "✅ Successfully pushed to https://github.com/nikhilanandd/Infrawatch/tree/$BRANCH"
echo ""
echo "📋 Next steps:"
echo "   1. Go to https://github.com/nikhilanandd/Infrawatch"
echo "   2. You'll see the 'dev' branch with all commits"
echo "   3. Do NOT merge to main yet — frontend is missing"
echo "   4. Create a PR when ready: dev -> main"