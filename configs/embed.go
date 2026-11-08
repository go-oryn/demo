// Package configs embeds the configuration files: the root ones, and one directory per environment (APP_ENV).
package configs

import "embed"

// FS holds the configuration files.
//
//go:embed *.yaml */*.yaml
var FS embed.FS
