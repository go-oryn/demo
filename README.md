# oryn demo

A small but complete [oryn](https://github.com/go-oryn/oryn) application: a tasks API on a database, with migrations, health checks, and traces, metrics and logs.

See the [demo documentation](https://go-oryn.github.io/oryn/demo/) for a guided tour.

## Run

```shell
go run .
```

It uses a SQLite database file (`demo.db`) and prints readable logs:

```shell
curl -X POST localhost:8080/api/tasks -d '{"title": "Try oryn"}'
curl localhost:8080/api/tasks
curl -X POST localhost:8080/api/tasks/<id>/complete
curl -X DELETE localhost:8080/api/tasks/<id>
curl 'localhost:8081/readyz?verbose'
```

Every response has an `X-Trace-Id` header.

## Telemetry

| Command | Telemetry |
|---|---|
| `go run .` | Logs on the console |
| `APP_ENV=stdout go run .` | Traces, metrics and logs printed on stdout |
| `docker compose up -d` then `APP_ENV=compose go run .` | PostgreSQL, and OTLP to Grafana on [localhost:3000](http://localhost:3000) |

## Layout

| Path | Content |
|---|---|
| `main.go` | Runs the application: migrations, health server (`:8081`), API server (`:8080`) |
| `app.go` | Assembles the application from the oryn modules |
| `configs/` | Configuration, per `APP_ENV` |
| `migrations/` | SQL migrations, embedded |
| `tasks/` | Repository, handlers, middlewares, and their module |
| `app_test.go` | Tests of the whole application, in memory |

## Test

```shell
go test ./...
```
