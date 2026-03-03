package websocket

import (
	"log/slog"
	"os"
	"testing"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

func TestNewHub(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	hub := NewHub(logger)
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
}

func TestHubRunAndBroadcast(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	hub := NewHub(logger)

	// Run hub in background
	go hub.Run()

	// Broadcast with no clients should not panic
	msg := models.WSMessage{
		Type:    "metrics",
		Payload: "test",
	}
	hub.Broadcast(msg)
}
