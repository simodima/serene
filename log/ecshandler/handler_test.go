package ecshandler

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/simodima/serene/log"
)

func TestECSHandler(t *testing.T) {
	var buf bytes.Buffer
	outputHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{})

	// Custom ECSHandler that wraps the buffered JSONHandler
	handler := &ECSHandler{
		JSONHandler:  outputHandler,
		levelRenamer: func(level slog.Level) string { return "renamed_" + strings.ToLower(level.String()) },
	}

	ctx := context.Background()
	ctx = log.AddLabelAttrs(ctx, slog.String("custom_key", "custom_value"))

	// Prepare: Create a sample slog.Record
	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelError,
		Message: "Test message",
		PC:      getCurrentPC(),
	}

	// Inject ECS-compliant attributes and call ECSHandler's Handle method
	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse the result into a JSON object for validation
	var logOutput map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logOutput); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	expectedKeys := map[string]any{
		ecsVersionKey: ecsVersion,
		timestampKey:  nil, // Just check existence for dynamic fields
		messageKey:    "Test message",
		logLevelKey:   "renamed_error",
		logLoggerKey:  logger,
		labelsKey:     nil,
		logOriginKey:  nil,
	}

	for key, expectedValue := range expectedKeys {
		value, ok := logOutput[key]
		if !ok {
			t.Errorf("missing %q key in log output", key)
			continue
		}

		// If a specific value is expected, check it
		if expectedValue != nil && value != expectedValue {
			t.Errorf("unexpected %q value: got %v, want %v", key, value, expectedValue)
		}
	}

	// Validate: Check that contextual attributes (from GetAttrsCtx) are included under "labels"
	if labels, ok := logOutput[labelsKey].(map[string]interface{}); !ok {
		t.Errorf("missing %q key or invalid value in log output", labelsKey)
	} else if labels["custom_key"] != "custom_value" {
		t.Errorf("unexpected %q value: got %v, want %q", "custom_key", labels["custom_key"], "custom_value")
	}

	// Check that log origin details are present
	if origin, ok := logOutput[logOriginKey].(map[string]interface{}); !ok {
		t.Errorf("missing %q key or invalid value in log output", logOriginKey)
	} else {
		// These are dynamically generated; confirm their presence but don't hardcode values
		requiredKeys := []string{fileNameKey, fileLineKey, functionKey}
		for _, rk := range requiredKeys {
			if _, ok := origin[rk]; !ok {
				t.Errorf("missing %q key in log origin", rk)
			}
		}
	}
}

// Helper function to mock retrieving a function's PC (Program Counter) for testing
func getCurrentPC() uintptr {
	pc, _, _, _ := runtime.Caller(1)
	return pc
}
