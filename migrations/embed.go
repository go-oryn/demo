// Package migrations embeds the database migrations, applied when the application starts.
package migrations

import "embed"

// FS holds the goose SQL migrations.
//
//go:embed *.sql
var FS embed.FS
