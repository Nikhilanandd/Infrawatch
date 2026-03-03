# ============================================
# InfraWatch Makefile
# ============================================

.PHONY: all build build-server build-agent build-web clean test run-server run-agent certs docker docker-up docker-down lint

BINARY_SERVER := bin/infrawatch-server
BINARY_AGENT  := bin/infrawatch-agent
GO            := go
GOFLAGS       := -ldflags="-s -w"

all: build

# ---- Build ----

build: build-server build-agent

build-server:
	@echo "=> Building server..."
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(BINARY_SERVER) ./cmd/server
	@echo "=> $(BINARY_SERVER) built"

build-agent:
	@echo "=> Building agent..."
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o $(BINARY_AGENT) ./cmd/agent
	@echo "=> $(BINARY_AGENT) built"

build-web:
	@echo "=> Building frontend..."
	cd web && npm install --legacy-peer-deps && npm run build
	@echo "=> Frontend built"

# ---- Run ----

run-server: build-server
	./$(BINARY_SERVER) --config configs/server.yaml

run-agent: build-agent
	./$(BINARY_AGENT) --config configs/agent.yaml

# ---- Test ----

test:
	$(GO) test ./... -v -race -count=1

# ---- Lint ----

lint:
	golangci-lint run ./...

# ---- Docker ----

docker:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# ---- Certs ----

certs:
	@echo "=> Generating self-signed mTLS certificates..."
	@bash scripts/gen-certs.sh
	@echo "=> Certificates generated in certs/"

# ---- Clean ----

clean:
	rm -rf bin/
	rm -rf web/build/
	rm -f infrawatch.db

# ---- Dependencies ----

deps:
	$(GO) mod tidy
	$(GO) mod download

# ---- Help ----

help:
	@echo "InfraWatch - Infrastructure Monitoring System"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build server and agent binaries"
	@echo "  build-server   Build server binary"
	@echo "  build-agent    Build agent binary"
	@echo "  build-web      Build React frontend"
	@echo "  run-server     Build and run server"
	@echo "  run-agent      Build and run agent"
	@echo "  test           Run tests"
	@echo "  lint           Run linter"
	@echo "  docker         Build Docker images"
	@echo "  docker-up      Start with Docker Compose"
	@echo "  docker-down    Stop Docker Compose"
	@echo "  certs          Generate self-signed mTLS certs"
	@echo "  clean          Remove build artifacts"
	@echo "  deps           Download Go dependencies"
