package tasks

import (
	"go.uber.org/fx"

	"github.com/go-oryn/oryn/httpserver"
)

// Module registers the repository, the middlewares and the API handlers.
var Module = fx.Options(
	fx.Provide(NewRepository),

	// Global middleware: applied to every handler.
	httpserver.AsMiddleware(TraceIDMiddleware),

	// The API handlers, under /api, with their own middleware.
	httpserver.AsHandlerGroup("/api",
		httpserver.WithMiddlewares(JSONMiddleware),
		httpserver.WithHandler("GET /tasks", NewListHandler),
		httpserver.WithHandler("POST /tasks", NewCreateHandler),
		httpserver.WithHandler("GET /tasks/{id}", NewGetHandler),
		httpserver.WithHandler("POST /tasks/{id}/complete", NewCompleteHandler),
		httpserver.WithHandler("DELETE /tasks/{id}", NewDeleteHandler),
	),
)
