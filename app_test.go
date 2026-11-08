package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/fx"

	"github.com/go-oryn/oryn/core/config"
	"github.com/go-oryn/oryn/core/coretest"
	"github.com/go-oryn/oryn/core/otel/oteltest/logtest"
	"github.com/go-oryn/oryn/core/otel/oteltest/metrictest"
	"github.com/go-oryn/oryn/core/otel/oteltest/spantest"
	"github.com/go-oryn/oryn/sql"

	"github.com/go-oryn/demo/tasks"
)

// testApp is the demo application running in a test, with its in-memory telemetry.
type testApp struct {
	t       *testing.T
	server  *http.Server
	spans   *tracetest.InMemoryExporter
	logs    *logtest.InMemoryExporter
	metrics *sdkmetric.ManualReader
}

// startApp starts the application with a SQLite database of its own, migrated.
func startApp(t *testing.T) *testApp {
	t.Helper()

	app := &testApp{t: t}

	coretest.Start(t, App,
		config.AsOptions(config.WithValues(map[string]any{
			"sql.driver": "sqlite",
			"sql.dsn":    filepath.Join(t.TempDir(), "test.db"),
		})),
		sql.Migrate(),
		fx.Populate(&app.server, &app.spans, &app.logs, &app.metrics),
	)

	return app
}

// do serves a request, without network.
func (a *testApp) do(method string, target string, body string) *httptest.ResponseRecorder {
	a.t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(a.t.Context(), method, target, strings.NewReader(body))
	a.server.Handler.ServeHTTP(recorder, request)

	return recorder
}

// create creates a task, and returns it.
func (a *testApp) create(title string) tasks.Task {
	a.t.Helper()

	response := a.do(http.MethodPost, "/api/tasks", `{"title": "`+title+`"}`)
	require.Equal(a.t, http.StatusCreated, response.Code, response.Body.String())

	var task tasks.Task
	require.NoError(a.t, json.Unmarshal(response.Body.Bytes(), &task))

	return task
}

func TestCreateTask(t *testing.T) {
	t.Parallel()

	app := startApp(t)

	response := app.do(http.MethodPost, "/api/tasks", `{"title": "test_title"}`)
	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
	assert.Len(t, response.Header().Get("X-Trace-Id"), 32)

	var task tasks.Task
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &task))
	assert.Equal(t, "test_title", task.Title)
	assert.False(t, task.Done)

	// The request span, with the task ID, and the INSERT query span in the same trace.
	spantest.HasSpan(t, app.spans,
		spantest.WithName("POST /api/tasks"),
		spantest.WithAttributeValue(attribute.String("task.id", task.ID)),
	)
	spantest.HasSpan(t, app.spans, spantest.WithName("INSERT"))
	logtest.HasRecord(t, app.logs, logtest.WithBody("task created"), logtest.WithLevel(slog.LevelInfo),
		logtest.WithAttributeValue("task_id", task.ID))
	assert.Equal(t, int64(1), metrictest.Int64Sum(t, app.metrics, "tasks.created"))
}

func TestCreateTaskInvalid(t *testing.T) {
	t.Parallel()

	app := startApp(t)

	for _, body := range []string{"", "not json", `{"title": "  "}`} {
		response := app.do(http.MethodPost, "/api/tasks", body)
		assert.Equal(t, http.StatusBadRequest, response.Code, body)
	}
}

func TestTaskLifecycle(t *testing.T) {
	t.Parallel()

	app := startApp(t)

	first := app.create("first")
	second := app.create("second")

	// List.
	response := app.do(http.MethodGet, "/api/tasks", "")
	require.Equal(t, http.StatusOK, response.Code)

	var list []tasks.Task
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
	require.Len(t, list, 2)
	assert.Equal(t, first.ID, list[0].ID)
	assert.Equal(t, second.ID, list[1].ID)

	// Complete, then get.
	assert.Equal(t, http.StatusNoContent, app.do(http.MethodPost, "/api/tasks/"+first.ID+"/complete", "").Code)

	response = app.do(http.MethodGet, "/api/tasks/"+first.ID, "")
	require.Equal(t, http.StatusOK, response.Code)

	var task tasks.Task
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &task))
	assert.True(t, task.Done)
	assert.Equal(t, first.CreatedAt, task.CreatedAt)

	// Delete.
	assert.Equal(t, http.StatusNoContent, app.do(http.MethodDelete, "/api/tasks/"+first.ID, "").Code)
	assert.Equal(t, http.StatusNotFound, app.do(http.MethodGet, "/api/tasks/"+first.ID, "").Code)
}

func TestTaskNotFound(t *testing.T) {
	t.Parallel()

	app := startApp(t)

	for _, request := range []struct{ method, target string }{
		{http.MethodGet, "/api/tasks/missing"},
		{http.MethodPost, "/api/tasks/missing/complete"},
		{http.MethodDelete, "/api/tasks/missing"},
	} {
		response := app.do(request.method, request.target, "")
		assert.Equal(t, http.StatusNotFound, response.Code, request.target)
		assert.JSONEq(t, `{"error": "task not found"}`, response.Body.String(), request.target)
	}
}

func TestEmptyList(t *testing.T) {
	t.Parallel()

	app := startApp(t)

	response := app.do(http.MethodGet, "/api/tasks", "")
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[]`, response.Body.String())
}
