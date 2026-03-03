package logger

import (
	"io"
	"log/slog"
	"os"
)

// New creates a structured JSON logger.
func New(level string, w ...io.Writer) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	var writer io.Writer = os.Stdout
	if len(w) > 0 && w[0] != nil {
		writer = w[0]
	}

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level:     l,
		AddSource: true,
	})
	return slog.New(handler)
}
