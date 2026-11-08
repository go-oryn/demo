// Command serve runs the demo servers in one process: the HTTP API (:8080), the gRPC API (:50051),
// the MCP server (:8082) and the health checks (:8081). It runs until SIGINT or SIGTERM, then
// stops them gracefully.
//
//	go run ./cmd/serve
package main

import (
	"github.com/go-oryn/oryn/core/health"
	"github.com/go-oryn/oryn/grpcserver"
	"github.com/go-oryn/oryn/httpserver"
	"github.com/go-oryn/oryn/mcpserver"

	"github.com/go-oryn/demo"
)

func main() {
	demo.App.Run(
		health.Serve(),
		httpserver.Serve(),
		grpcserver.Serve(),
		mcpserver.Serve(),
	)
}
