package tasks

import "go.uber.org/fx"

// Module provides the tasks [Service], and its [Repository]. The API modules
// (httpapi, grpcapi, mcpapi) depend on it.
var Module = fx.Module("tasks",
	fx.Provide(NewRepository, NewService),
)
