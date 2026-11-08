// Package demotest runs the demo application in tests, in memory: a SQLite database of its own,
// and clients for the HTTP, gRPC and MCP APIs, without network.
package demotest

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-oryn/oryn/core/config"
	"github.com/go-oryn/oryn/core/coretest"
	"github.com/go-oryn/oryn/core/otel/oteltest/logtest"
	"github.com/go-oryn/oryn/grpcserver/grpcservertest"
	"github.com/go-oryn/oryn/mcpserver/mcpservertest"
	"github.com/go-oryn/oryn/sql"

	"github.com/go-oryn/demo"
	tasksv1 "github.com/go-oryn/demo/proto/tasks/v1"
)

// App is the demo application started in a test, with API clients and its in-memory telemetry.
type App struct {
	t *testing.T

	// HTTPServer is the HTTP server: send it requests with [App.HTTP].
	HTTPServer *http.Server
	// GRPC is a client of the gRPC TasksService, connected in memory.
	GRPC tasksv1.TasksServiceClient
	// MCP is an MCP client session, connected in memory.
	MCP *mcp.ClientSession

	Spans   *tracetest.InMemoryExporter
	Logs    *logtest.InMemoryExporter
	Metrics *sdkmetric.ManualReader
}

// Start starts the demo application with a migrated SQLite database of its own. It is stopped
// at the end of the test. The options are added to the application, e.g. fx.Populate.
func Start(t *testing.T, options ...fx.Option) *App {
	t.Helper()

	app := &App{t: t}

	var (
		grpcServer *grpc.Server
		mcpServer  *mcp.Server
	)

	coretest.Start(t, demo.App,
		append(options,
			config.AsOptions(config.WithValues(map[string]any{
				"sql.driver": "sqlite",
				"sql.dsn":    filepath.Join(t.TempDir(), "test.db"),
			})),
			sql.Migrate(),
			fx.Populate(&app.HTTPServer, &grpcServer, &mcpServer, &app.Spans, &app.Logs, &app.Metrics),
		)...,
	)

	app.GRPC = tasksv1.NewTasksServiceClient(grpcservertest.Dial(t, grpcServer))
	app.MCP = mcpservertest.Connect(t, mcpServer)

	return app
}

// HTTP serves an HTTP request, without network, and returns the response.
func (a *App) HTTP(method string, target string, body string) *httptest.ResponseRecorder {
	a.t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(a.t.Context(), method, target, strings.NewReader(body))
	a.HTTPServer.Handler.ServeHTTP(recorder, request)

	return recorder
}
