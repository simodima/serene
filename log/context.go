package log

import (
	"context"
	"log/slog"
	"net/http"
)

// slogAttributes is the attributes context key type
type slogAttributes string

var (
	labelsAttributes = slogAttributes("labels")
	ecsAttributes    = slogAttributes("ecs")
)

// LogExtract is a function extract informations from log
type LogExtract func(r *http.Request) (slog.Attr, bool)

func ExtractHeaderRename(name string, rename string) LogExtract {
	return func(r *http.Request) (slog.Attr, bool) {
		if val := r.Header.Get(name); len(val) != 0 {
			return slog.String(rename, val), true
		}
		return slog.Attr{}, false
	}
}

// AddLabelAttrs appends the given slog attributes to the context.
func AddLabelAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	attrs = append(attrs, GetLabelAttrs(ctx)...)
	return context.WithValue(ctx, labelsAttributes, attrs)
}

// addECSAttrs appends the given slog attributes to the context.
func addECSAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	attrs = append(attrs, GetECSAttrs(ctx)...)
	return context.WithValue(ctx, ecsAttributes, attrs)
}

// GetLabelAttrs gets the slog attributes from the give context.
func GetLabelAttrs(ctx context.Context) []slog.Attr {
	if loadedAttrs, ok := ctx.Value(labelsAttributes).([]slog.Attr); ok {
		return loadedAttrs
	}
	return []slog.Attr{}
}

// GetECSAttrs gets the slog attributes from the give context.
func GetECSAttrs(ctx context.Context) []slog.Attr {
	if loadedAttrs, ok := ctx.Value(ecsAttributes).([]slog.Attr); ok {
		return loadedAttrs
	}
	return []slog.Attr{}
}

type middlewareOptions struct {
	labelExtractors []LogExtract
	ecsExtractors   []LogExtract
	logRequest      bool
}

type MiddlewareOption func(*middlewareOptions)

func LogRequest() MiddlewareOption {
	return func(o *middlewareOptions) {
		o.logRequest = true
	}
}

func WithDefaultInfo() MiddlewareOption {
	return func(o *middlewareOptions) {
		o.ecsExtractors = append(
			o.ecsExtractors,
			func(r *http.Request) (slog.Attr, bool) {
				return slog.String("http.request.method", r.Method), true
			},
		)
	}
}

// HTTPAttributesMiddleware is an HTTP middleware
// for associating HTTP data with structured logging attributes.
// By inspecting specified request headers and transforming them into structured logging attributes,
// this middleware can enhance the logging capabilities of an application by attaching
// contextual data from HTTP request headers.
func HTTPAttributesMiddleware(opts ...MiddlewareOption) func(http.Handler) http.Handler {
	options := middlewareOptions{}

	for _, o := range opts {
		o(&options)
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ecsAttributes := []slog.Attr{}
			for _, extract := range options.ecsExtractors {
				if attr, ok := extract(r); ok {
					ecsAttributes = append(ecsAttributes, attr)
				}
			}
			ctx := addECSAttrs(r.Context(), ecsAttributes...)

			labelAttributes := []slog.Attr{}
			for _, extract := range options.labelExtractors {
				if attr, ok := extract(r); ok {
					labelAttributes = append(labelAttributes, attr)
				}
			}
			ctx = AddLabelAttrs(ctx, labelAttributes...)

			h.ServeHTTP(w, r.WithContext(ctx))

			if options.logRequest {
				slog.InfoContext(ctx, "HTTP Request handled")
			}
		})
	}
}
