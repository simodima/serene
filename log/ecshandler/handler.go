package ecshandler

import (
	"context"
	"log/slog"
	"os"
	"runtime"

	"github.com/simodima/serene/log"
)

// Package constants for ECS (Elastic Common Schema) and logger metadata
const (
	ecsVersion = "8.11.0"   // ECS version being used
	logger     = "log/slog" // Logger name
)

// Key constants used in log fields for ECS compliance
const (
	ecsVersionKey = "ecs.version" // ECS version key

	timestampKey = "@timestamp" // ECS-compliant timestamp key
	messageKey   = "message"    // Message content key
	logLevelKey  = "log.level"  // Logging level key
	logLoggerKey = "log.logger" // Logger name key
	fileNameKey  = "file.name"  // Source file name key
	fileLineKey  = "file.line"  // Source file line number key
	logOriginKey = "log.origin" // Log origin group key
	functionKey  = "function"   // Function name key
	labelsKey    = "labels"     // Labels group key

	// Error keys (currently commented out)
	// errorKey           = "error"
	// errorMessageKey    = "message"
	// errorStackTraceKey = "stack_trace"
)

// Option represents a functional option to customize the ECSHandler.
type Option func(*opts)

// opts is a struct that holds configuration options for the ECSHandler.
type opts struct {
	level        slog.Level                                   // Minimum logging level
	levelRenamer func(slog.Level) string                      // Function to rename logging levels
	replaceAttr  func(groups []string, a slog.Attr) slog.Attr // Function to replace/modify attributes
}

// defaultOptions defines the default settings for the ECSHandler.
var defaultOptions = opts{
	level: slog.LevelDebug, // Default logging level: Debug
	levelRenamer: func(level slog.Level) string { // Default level renamer: uses level's string representation
		return level.String()
	},
	// Default ReplaceAttr function removes certain automatically added attributes (e.g., "time", "msg", etc.)
	replaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case "time", "msg", "source", "level":
			return slog.Attr{} // Exclude these attributes
		default:
			return a // Keep all other attributes
		}
	},
}

// WithLevel sets the minimum logging level for the ECSHandler.
func WithLevel(l slog.Level) Option {
	return func(o *opts) {
		o.level = l
	}
}

// WithLevelRenamer sets a custom function to rename or transform log levels.
func WithLevelRenamer(fn func(slog.Level) string) Option {
	return func(o *opts) {
		o.levelRenamer = fn
	}
}

// WithReplaceAttr sets a custom function to replace or modify attributes in log records.
func WithReplaceAttr(fn func(groups []string, a slog.Attr) slog.Attr) Option {
	return func(o *opts) {
		o.replaceAttr = fn
	}
}

// ECSHandler is a log handler that formats logs following the ECS (Elastic Common Schema).
type ECSHandler struct {
	*slog.JSONHandler                         // Underlying handler for writing JSON-formatted logs
	levelRenamer      func(slog.Level) string // Custom function for renaming levels
}

// NewECSHandler creates a new ECSHandler with the specified options.
// It wraps a JSONHandler and enforces ECS-compliant structured logging.
func NewECSHandler(options ...Option) *ECSHandler {
	o := defaultOptions // Start with default options
	for _, op := range options {
		op(&o) // Apply each Option to override defaults
	}

	// Create a JSONHandler with the configured options
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       o.level,       // Set log level
		ReplaceAttr: o.replaceAttr, // Set attribute replacement function
	})

	// Return an ECSHandler wrapping the JSONHandler
	return &ECSHandler{
		JSONHandler:  h,
		levelRenamer: o.levelRenamer,
	}
}

// Handle processes a log record (slog.Record) and transforms it to match ECS requirements.
// Adds ECS-compliant fields and passes the record to the JSONHandler.
func (h *ECSHandler) Handle(ctx context.Context, r slog.Record) error {
	labels := log.GetLabelAttrs(ctx)
	plainAttributes := log.GetECSAttrs(ctx)

	// Obtain stack frame information (e.g., file, line, function) for the log origin
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next() // Retrieve the next frame

	allAttributes := append([]slog.Attr{
		slog.Time(timestampKey, r.Time),                     // Add log timestamp
		slog.String(messageKey, r.Message),                  // Add log message
		slog.String(logLevelKey, h.levelRenamer(r.Level)),   // Add log level (renamed if necessary)
		slog.String(ecsVersionKey, ecsVersion),              // Add ECS version
		slog.String(logLoggerKey, logger),                   // Add logger name
		{Key: labelsKey, Value: slog.GroupValue(labels...)}, // Add contextual attributes as "labels"

		// Add log origin details (file, line, function)
		slog.Group(logOriginKey,
			slog.String(fileNameKey, f.File),     // File name
			slog.Int(fileLineKey, f.Line),        // Line number
			slog.String(functionKey, f.Function), // Function name
		),
	}, plainAttributes...)

	// Add ECS-compliant attributes to the log record
	r.AddAttrs(allAttributes...)

	// Pass the enriched log record to the underlying JSONHandler
	return h.JSONHandler.Handle(ctx, r)
}
