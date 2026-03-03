package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikhilanandd/infrawatch/internal/config"
	"github.com/nikhilanandd/infrawatch/internal/server/alert"
	"github.com/nikhilanandd/infrawatch/internal/server/api"
	"github.com/nikhilanandd/infrawatch/internal/server/store"
	"github.com/nikhilanandd/infrawatch/internal/server/websocket"
	"github.com/nikhilanandd/infrawatch/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfgPath := flag.String("config", "configs/server.yaml", "path to server config")
	flag.Parse()

	cfg, err := config.LoadServerConfig(*cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Server.LogLevel)
	log.Info("starting infrawatch server", "port", cfg.Server.Port)

	// Initialize database
	db, err := store.New(cfg.Database.Path)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Seed default admin user if not exists
	seedDefaultUser(db, log)

	// Start WebSocket hub
	hub := websocket.NewHub(log)
	go hub.Run()

	// Start alert engine
	alertEngine := alert.NewEngine(db, hub, cfg, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go alertEngine.Run(ctx)

	// Setup API handler
	handler := &api.Handler{
		DB:     db,
		Hub:    hub,
		Config: cfg,
		Logger: log,
	}

	router := handler.NewRouter()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Configure mTLS if enabled
	if cfg.Server.TLS.Enabled {
		tlsConfig, err := setupTLS(cfg)
		if err != nil {
			log.Error("failed to setup TLS", "error", err)
			os.Exit(1)
		}
		srv.TLSConfig = tlsConfig
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Info("received signal, shutting down", "signal", sig)
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("server shutdown error", "error", err)
		}
	}()

	// Start server
	log.Info("server listening", "addr", addr, "tls", cfg.Server.TLS.Enabled)

	if cfg.Server.TLS.Enabled {
		err = srv.ListenAndServeTLS(cfg.Server.TLS.ServerCert, cfg.Server.TLS.ServerKey)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped")
}

func setupTLS(cfg *config.ServerConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.Server.TLS.CACert != "" {
		caCert, err := os.ReadFile(cfg.Server.TLS.CACert)
		if err != nil {
			return nil, fmt.Errorf("read CA cert: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert")
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
	}

	return tlsConfig, nil
}

func seedDefaultUser(db *store.DB, log *slog.Logger) {
	_, err := db.GetUserByUsername("admin")
	if err == nil {
		return // admin already exists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash default password", "error", err)
		return
	}

	if err := db.CreateUser("admin", string(hash), "admin"); err != nil {
		log.Error("failed to create default admin", "error", err)
		return
	}

	log.Info("created default admin user", "username", "admin", "password", "admin")
}
