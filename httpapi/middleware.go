package httpapi

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// TraceIDMiddleware adds the X-Trace-Id header to the responses: the ID of the request trace,
// to find it in the observability backend.
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if spanContext := trace.SpanContextFromContext(r.Context()); spanContext.HasTraceID() {
			w.Header().Set("X-Trace-Id", spanContext.TraceID().String())
		}

		next.ServeHTTP(w, r)
	})
}

// JSONMiddleware sets the JSON content type of the API responses.
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}
