// Package httpapi (internal/http) is the HTTP API of the tasks: JSON handlers on the [tasks.Service].
// It is not named http, so it does not collide with net/http.
package httpapi

import (
	"go.uber.org/fx"

	"github.com/go-oryn/oryn/httpserver"
)

// Module registers the HTTP middlewares and handlers.
var Module = fx.Module("httpapi",
	// A global middleware, applied to every handler...
	httpserver.AsMiddleware(TraceIDMiddleware),

	// ...and the API handlers, under /api, with their own middleware.
	httpserver.AsHandlerGroup("/api",
		httpserver.WithMiddlewares(JSONMiddleware),
		httpserver.WithHandler("GET /tasks", NewListHandler),
		httpserver.WithHandler("POST /tasks", NewCreateHandler),
		httpserver.WithHandler("GET /tasks/{id}", NewGetHandler),
		httpserver.WithHandler("POST /tasks/{id}/complete", NewCompleteHandler),
		httpserver.WithHandler("DELETE /tasks/{id}", NewDeleteHandler),
	),
)
