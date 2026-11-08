# Agents guide

This is an [oryn](https://github.com/go-oryn/oryn) application. Read https://go-oryn.github.io/oryn/llms.txt before writing code.

- Business rules and their telemetry live in the domain module, `tasks/` (`Service`, on the `Repository`).
- The APIs (`httpapi/`, `grpcapi/`, `mcpapi/`) only map requests and errors to the `tasks.Service`: keep them thin, and keep the three in sync.
- Register handlers with `httpserver.AsHandler` / `AsHandlerGroup`, gRPC services with `grpcserver.AsService`, MCP tools with `mcpserver.AsCapability`. Register new modules in `app.go`.
- Change the gRPC API in `proto/tasks/v1/tasks.proto`, then run `go generate ./proto` (needs protoc).
- Add schema changes as new goose files in `migrations/` (`00002_xxx.sql`), never edit applied ones. SQL uses `$1` placeholders: it must run on SQLite and PostgreSQL.
- Get the tracer, meter and logger from the fx container (`otel.Telemetry`), never the OpenTelemetry globals. Log with the `...Context` methods.
- Test with `demotest.Start`: it gives in-memory HTTP, gRPC and MCP clients, and the in-memory telemetry to assert with the `oteltest` packages.

Checks:

```shell
go vet ./...
go test -race ./...
```
