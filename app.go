// Package demo is the oryn demo application: the tasks domain on a database, exposed as an
// HTTP API, a gRPC API and MCP tools. The commands of cmd/ run it.
package demo

import (
	"github.com/go-oryn/oryn/core"
	"github.com/go-oryn/oryn/grpcserver"
	"github.com/go-oryn/oryn/httpserver"
	"github.com/go-oryn/oryn/mcpserver"
	"github.com/go-oryn/oryn/sql"

	"github.com/go-oryn/demo/configs"
	"github.com/go-oryn/demo/grpcapi"
	"github.com/go-oryn/demo/httpapi"
	"github.com/go-oryn/demo/mcpapi"
	"github.com/go-oryn/demo/migrations"
	"github.com/go-oryn/demo/tasks"
)

// App is the demo application: the modules it uses, and its own code. The commands of cmd/
// run it with different options, and the tests reuse it.
var App = core.New(
	core.WithConfigFS(configs.FS),
	httpserver.Module,
	grpcserver.Module,
	mcpserver.Module,
	sql.Module,
	sql.AsMigrations(migrations.FS),
	// The tasks domain, and the three APIs on it.
	tasks.Module,
	httpapi.Module,
	grpcapi.Module,
	mcpapi.Module,
)
