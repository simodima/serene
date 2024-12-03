package log

import (
	"context"
	"log/slog"
	"os"
	"runtime"
)

const (
	ecsVersion = "8.11.0"
	logger     = "log/slog"
)

const (
	ecsVersionKey = "ecs.version"

	timestampKey = "@timestamp"
	messageKey   = "message"
	logLevelKey  = "log.level"
	logLoggerKey = "log.logger"
	fileNameKey  = "file.name"
	fileLineKey  = "file.line"
	logOriginKey = "log.origin"
	functionKey  = "function"
	labelsKey    = "labels"

	// errorKey           = "error"
	// errorMessageKey    = "message"
	// errorStackTraceKey = "stack_trace"
)

type Option func(*opts)

type opts struct {
	level        slog.Level
	levelRenamer func(slog.Level) string
	replaceAttr  func(groups []string, a slog.Attr) slog.Attr
}

var defaultOptions = opts{
	level:        slog.LevelDebug,
	levelRenamer: func(level slog.Level) string { return level.String() },
	replaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case "time", "msg", "source", "level":
			return slog.Attr{}
		default:
			return a
		}
	},
}

func WithLevel(l slog.Level) Option {
	return func(o *opts) {
		o.level = l
	}
}

func WithLevelRenamer(fn func(slog.Level) string) Option {
	return func(o *opts) {
		o.levelRenamer = fn
	}
}

func WithReplaceAttr(fn func(groups []string, a slog.Attr) slog.Attr) Option {
	return func(o *opts) {
		o.replaceAttr = fn
	}
}

type ECSHandler struct {
	*slog.JSONHandler
	levelRenamer func(slog.Level) string
}

func NewECSHandler(options ...Option) *ECSHandler {
	o := defaultOptions
	for _, op := range options {
		op(&o)
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       o.level,
		ReplaceAttr: o.replaceAttr,
	})

	return &ECSHandler{
		JSONHandler:  h,
		levelRenamer: o.levelRenamer,
	}
}

func (h *ECSHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := GetAttrsCtx(ctx)
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()

	r.AddAttrs(
		slog.Time(timestampKey, r.Time),
		slog.String(messageKey, r.Message),
		slog.String(logLevelKey, h.levelRenamer(r.Level)),
		slog.String(ecsVersionKey, ecsVersion),
		slog.String(logLoggerKey, logger),
		slog.Attr{Key: labelsKey, Value: slog.GroupValue(attrs...)},
		slog.Group(logOriginKey,
			slog.String(fileNameKey, f.File),
			slog.Int(fileLineKey, f.Line),
			slog.String(functionKey, f.Function),
		),
	)

	return h.JSONHandler.Handle(ctx, r)
}
