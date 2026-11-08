package tasks

import "go.uber.org/fx"

// Module provides the tasks [Service], on the [Repository] as [Store]. The API modules
// (internal/http, internal/grpc, internal/mcp) depend on it.
var Module = fx.Module("tasks",
	fx.Provide(
		fx.Annotate(NewRepository, fx.As(new(Store))),
		NewService,
	),
)
