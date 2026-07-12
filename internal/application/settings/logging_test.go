package settings_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
)

func TestLoggerWithSystemSettingsHostnameOnly(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger := settings.LoggerWithSystemSettings(base, settings.SystemSettings{
		Timezone: "UTC",
		Hostname: "vyntrio",
	}, false)
	logger.Info("test message")

	attrs := parseLogAttrs(t, buf.String())
	if attrs["hostname"] != "vyntrio" {
		t.Fatalf("hostname = %v, want vyntrio", attrs["hostname"])
	}
	if _, ok := attrs["timezone"]; ok {
		t.Fatalf("timezone should not be attached: %v", attrs)
	}
}

func TestLoggerWithSystemSettingsIncludesTimezoneWhenRequested(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	logger := settings.LoggerWithSystemSettings(base, settings.SystemSettings{
		Timezone: "UTC",
		Hostname: "vyntrio",
	}, true)
	logger.Info("test message")

	attrs := parseLogAttrs(t, buf.String())
	if attrs["hostname"] != "vyntrio" {
		t.Fatalf("hostname = %v, want vyntrio", attrs["hostname"])
	}
	if attrs["timezone"] != "UTC" {
		t.Fatalf("timezone = %v, want UTC", attrs["timezone"])
	}
}

func parseLogAttrs(t *testing.T, line string) map[string]any {
	t.Helper()

	line = strings.TrimSpace(line)
	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("unmarshal log line: %v", err)
	}
	return payload
}
