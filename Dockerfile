# ============================================
# InfraWatch Server - Multi-stage Dockerfile
# ============================================

# Stage 1: Build Go server
FROM golang:1.22-alpine AS go-builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o infrawatch-server ./cmd/server

# Stage 2: Build React frontend
FROM node:20-alpine AS web-builder
WORKDIR /build
COPY web/package.json web/package-lock.json* ./
RUN npm install --legacy-peer-deps
COPY web/ .
RUN npm run build

# Stage 3: Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

RUN addgroup -S infrawatch && adduser -S infrawatch -G infrawatch

WORKDIR /app

COPY --from=go-builder /build/infrawatch-server .
COPY --from=web-builder /build/build ./web/build
COPY configs/server.yaml ./configs/server.yaml

RUN mkdir -p /var/lib/infrawatch /etc/infrawatch/certs && \
    chown -R infrawatch:infrawatch /var/lib/infrawatch /etc/infrawatch /app

USER infrawatch

EXPOSE 8443

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8443/health || exit 1

ENTRYPOINT ["./infrawatch-server"]
CMD ["--config", "configs/server.yaml"]
