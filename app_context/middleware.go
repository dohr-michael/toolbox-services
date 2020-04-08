package app_context

import (
	"context"
	"net/http"
)

type ContextEnrichment func(context.Context) context.Context

func ContextMiddleware(fns ...ContextEnrichment) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			for _, fn := range fns {
				ctx = fn(ctx)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
