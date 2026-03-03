package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	l := New("info")
	if l == nil {
		t.Fatal("New returned nil")
	}
}

func TestNewWithWriter(t *testing.T) {
	var buf bytes.Buffer
	l := New("info", &buf)

	l.Info("hello world", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Errorf("Expected log output to contain 'hello world', got: %s", output)
	}
	if !strings.Contains(output, `"key"`) {
		t.Errorf("Expected log output to contain key, got: %s", output)
	}

	// Verify JSON format
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Log output should be valid JSON: %v", err)
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level    string
		logFunc  string
		expected bool
	}{
		{"debug", "debug", true},
		{"debug", "info", true},
		{"info", "debug", false},
		{"info", "info", true},
		{"info", "warn", true},
		{"warn", "info", false},
		{"warn", "warn", true},
		{"error", "warn", false},
		{"error", "error", true},
	}

	for _, tc := range tests {
		t.Run(tc.level+"_"+tc.logFunc, func(t *testing.T) {
			var buf bytes.Buffer
			l := New(tc.level, &buf)

			switch tc.logFunc {
			case "debug":
				l.Debug("test message")
			case "info":
				l.Info("test message")
			case "warn":
				l.Warn("test message")
			case "error":
				l.Error("test message")
			}

			hasOutput := buf.Len() > 0
			if hasOutput != tc.expected {
				t.Errorf("level=%s logFunc=%s: output=%v, want %v",
					tc.level, tc.logFunc, hasOutput, tc.expected)
			}
		})
	}
}

func TestNewDefaultLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New("invalid-level", &buf)

	l.Info("should appear")
	if buf.Len() == 0 {
		t.Error("Invalid level should default to info, info message should appear")
	}

	buf.Reset()
	l.Debug("should not appear")
	if buf.Len() != 0 {
		t.Error("Invalid level (defaulting to info) should suppress debug messages")
	}
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New("info", &buf)

	l.Info("structured", "user", "alice", "count", 42)

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}
	if entry["msg"] != "structured" {
		t.Errorf("msg = %v, want %q", entry["msg"], "structured")
	}
	if entry["user"] != "alice" {
		t.Errorf("user = %v, want %q", entry["user"], "alice")
	}
}
