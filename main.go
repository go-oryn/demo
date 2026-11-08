// Command demo is the oryn demo application: a tasks API on a database.
package main

import (
	"github.com/go-oryn/oryn/core/health"
	"github.com/go-oryn/oryn/httpserver"
	"github.com/go-oryn/oryn/sql"
)

func main() {
	// Apply the migrations, then start the health checks (:8081) and the API (:8080) servers.
	// It runs until SIGINT or SIGTERM, then stops gracefully.
	App.Run(
		sql.Migrate(),
		health.Serve(),
		httpserver.Serve(),
	)
}
