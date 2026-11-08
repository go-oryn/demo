package main

import (
	"github.com/go-oryn/oryn/core"
	"github.com/go-oryn/oryn/httpserver"
	"github.com/go-oryn/oryn/sql"

	"github.com/go-oryn/demo/configs"
	"github.com/go-oryn/demo/migrations"
	"github.com/go-oryn/demo/tasks"
)

// App is the demo application: the modules it uses, and its own code. The tests reuse it.
var App = core.New(
	core.WithConfigFS(configs.FS),
	httpserver.Module,
	sql.Module,
	sql.AsMigrations(migrations.FS),
	tasks.Module,
)
