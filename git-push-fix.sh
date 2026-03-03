#!/bin/bash
set -e

REPO_URL="git@github.com:nikhilanandd/Infrawatch.git"
BRANCH="dev"

echo "🔧 InfraWatch — Resuming Git Push from Commit 8"
echo "================================================="

git checkout "$BRANCH"

echo ""
echo "=== Commit 8: SQLite storage layer ==="
git add internal/server/store/
git commit -m "feat(server): implement SQLite storage with WAL mode

- Auto-migration: creates nodes, metrics, alerts, users, alert_rules tables
- Time-series indexed metrics storage with JSON payload
- CRUD operations for nodes, metrics, alerts, users
- Default admin user seeded on first run (admin/admin)
- WAL mode enabled for concurrent read performance
- Configurable data retention and query windowing"

echo ""
echo "=== Commit 9: JWT authentication and RBAC ==="
git add internal/server/auth/
git commit -m "feat(server): add JWT authentication with role-based access control

- JWT token generation with username, role, and 24h expiry
- Token validation and claims extraction
- bcrypt password hashing and verification
- AuthMiddleware: validates Bearer token on protected routes
- RoleMiddleware: restricts endpoints by role (admin/viewer)
- Proper 401/403 error responses"

echo ""
echo "=== Commit 10: Alert engine and notifiers ==="
git add internal/server/alert/
git commit -m "feat(server): implement alert rule engine with email and Slack notifiers

- Rule evaluation: CPU > 80%, Disk > 85%, Service down
- Alert deduplication and severity classification
- Email notifier via SMTP with configurable recipients
- Slack notifier via webhook URL with formatted messages
- Extensible rule and notifier architecture"

echo ""
echo "=== Commit 11: WebSocket hub for live updates ==="
git add internal/server/websocket/
git commit -m "feat(server): add WebSocket hub for real-time metric streaming

- Hub pattern with client registration/unregistration
- Broadcast metrics and alerts to all connected clients
- Thread-safe client management
- Client count tracking"

echo ""
echo "=== Commit 12: REST API handlers and router ==="
git add internal/server/api/
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
git add -A -- '*_test.go'
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
git add Dockerfile Dockerfile.agent docker-compose.yml 2>/dev/null || true
git add scripts/ 2>/dev/null || true
git add deployments/ 2>/dev/null || true
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
echo "================================================="
echo "📊 Commit Summary:"
git log --oneline
echo ""
echo "================================================="
echo "🚀 Pushing to origin/$BRANCH..."
git push -u origin "$BRANCH"

echo ""
echo "✅ Successfully pushed to https://github.com/nikhilanandd/Infrawatch/tree/$BRANCH"