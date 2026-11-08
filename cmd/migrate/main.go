// Command migrate applies the pending database migrations, then exits.
//
//	go run ./cmd/migrate
package main

import (
	"github.com/go-oryn/oryn/sql"

	"github.com/go-oryn/demo"
)

func main() {
	demo.App.RunOnce(sql.Migrate())
}
