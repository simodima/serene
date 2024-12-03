package log

import (
	"context"
	"log/slog"
	"net/http"
)

type requestAttrsKey struct{}

var ctxAttrs requestAttrsKey

func AddAttrsCtx(ctx context.Context, attrs ...slog.Attr) context.Context {
	attrs = append(attrs, GetAttrsCtx(ctx)...)
	return context.WithValue(ctx, ctxAttrs, attrs)
}

func GetAttrsCtx(ctx context.Context) []slog.Attr {
	if loadedAttrs, ok := ctx.Value(ctxAttrs).([]slog.Attr); ok {
		return loadedAttrs
	}
	return []slog.Attr{}
}

type HeaderAttrs map[string]func(string) slog.Attr

func Header(newName string) func(string) slog.Attr {
	return func(value string) slog.Attr {
		return slog.String(newName, value)
	}
}

func RequestHeadersMiddleware(headers map[string]func(string) slog.Attr) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logAttributes := []slog.Attr{}

			for header, toAttr := range headers {
				if val := r.Header.Get(header); len(val) != 0 {
					logAttributes = append(logAttributes, toAttr(val))
				}
			}

			ctx := AddAttrsCtx(r.Context(), logAttributes...)

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
