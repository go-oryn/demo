# oryn demo

[![CI](https://github.com/go-oryn/demo/actions/workflows/ci.yml/badge.svg)](https://github.com/go-oryn/demo/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/go-oryn/demo/graph/badge.svg)](https://codecov.io/gh/go-oryn/demo)

A complete [oryn](https://github.com/go-oryn/oryn) application: one `tasks` module, served over an HTTP API, a gRPC API and MCP tools for AI agents, in one process, with its traces, metrics and logs.

Documentation: https://go-oryn.github.io/oryn/demo/overview/

## Run

```shell
go run ./cmd/migrate   # apply the migrations (SQLite demo.db)
go run ./cmd/serve     # HTTP :8080, gRPC :50051, MCP :8082, health :8081
```

```shell
# HTTP
curl -X POST localhost:8080/api/tasks -d '{"title": "Try oryn"}'
curl localhost:8080/api/tasks

# gRPC (reflection is enabled)
grpcurl -plaintext -d '{"title": "Try oryn"}' localhost:50051 tasks.v1.TasksService/CreateTask
grpcurl -plaintext localhost:50051 tasks.v1.TasksService/ListTasks

# MCP
claude mcp add --transport http oryn-demo http://localhost:8082/mcp
```

## Layout

| Path | Content |
|---|---|
| `app.go` | The application, shared by the commands and the tests |
| `cmd/serve`, `cmd/migrate` | The commands |
| `internal/tasks/` | The domain module: `Service` (rules, spans, metrics, logs) on a `Store`, the `Repository` (SQL) |
| `internal/http/`, `internal/grpc/`, `internal/mcp/` | The three APIs on the `tasks.Service` |
| `proto/` | The gRPC API definition and its generated code (`go generate ./proto`) |
| `migrations/` | SQL migrations, embedded |
| `configs/` | Configuration, per `APP_ENV` |
| `internal/demotest/` | Starts the application in tests, with in-memory HTTP, gRPC and MCP clients, and `fx.Replace` helpers |

## Telemetry

| Command | Telemetry |
|---|---|
| `go run ./cmd/serve` | Logs on the console |
| `APP_ENV=stdout go run ./cmd/serve` | Traces, metrics and logs printed on stdout |
| `docker compose up -d`, then `APP_ENV=compose go run ./cmd/migrate` and `APP_ENV=compose go run ./cmd/serve` | PostgreSQL, and OTLP to Grafana on [localhost:3000](http://localhost:3000) |

## Test

```shell
go test ./...
```
