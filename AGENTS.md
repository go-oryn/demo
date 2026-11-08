# Agents guide

This is an [oryn](https://github.com/go-oryn/oryn) application. The oryn documentation for agents is at https://go-oryn.github.io/oryn/llms.txt.

- Add features as fx modules, like `tasks/module.go`, and register them in `app.go`.
- Register handlers with `httpserver.AsHandler` / `httpserver.AsHandlerGroup`, middlewares with `httpserver.AsMiddleware`.
- Add schema changes as new goose files in `migrations/` (`00002_xxx.sql`), never edit applied ones.
- SQL uses `$1` placeholders: it must run on SQLite and PostgreSQL.
- Get the tracer, meter and logger from the fx container, never the OpenTelemetry globals.
- Test through the whole application with `coretest.Start` (see `app_test.go`), and assert telemetry with the `oteltest` packages.

Checks:

```shell
go vet ./...
go test -race ./...
```
