package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-oryn/oryn/core/otel/oteltest/spantest"

	"github.com/go-oryn/demo/demotest"
	"github.com/go-oryn/demo/tasks"
)

func TestHandlers(t *testing.T) {
	t.Parallel()

	t.Run("task lifecycle", func(t *testing.T) {
		t.Parallel()

		app := demotest.Start(t)

		// Create.
		response := app.HTTP(http.MethodPost, "/api/tasks", `{"title": "test_title"}`)
		require.Equal(t, http.StatusCreated, response.Code)
		assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
		assert.Len(t, response.Header().Get("X-Trace-Id"), 32)

		var task tasks.Task
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &task))
		assert.Equal(t, "test_title", task.Title)

		// Complete, then list.
		response = app.HTTP(http.MethodPost, "/api/tasks/"+task.ID+"/complete", "")
		require.Equal(t, http.StatusOK, response.Code)

		response = app.HTTP(http.MethodGet, "/api/tasks", "")
		require.Equal(t, http.StatusOK, response.Code)

		var list []tasks.Task
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
		require.Len(t, list, 1)
		assert.True(t, list[0].Done)

		// Delete, then get.
		assert.Equal(t, http.StatusNoContent, app.HTTP(http.MethodDelete, "/api/tasks/"+task.ID, "").Code)
		assert.Equal(t, http.StatusNotFound, app.HTTP(http.MethodGet, "/api/tasks/"+task.ID, "").Code)

		// The tasks.create span is a child of the HTTP server span.
		server := spantest.Find(app.Spans, spantest.WithName("POST /api/tasks"))
		require.Len(t, server, 1)
		spantest.HasSpan(t, app.Spans, spantest.WithName("tasks.create"), spantest.WithParent(server[0].SpanContext.SpanID()))
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		app := demotest.Start(t)

		for _, tc := range []struct {
			method, target, body string
			status               int
		}{
			{http.MethodPost, "/api/tasks", "not json", http.StatusBadRequest},
			{http.MethodPost, "/api/tasks", `{"title": " "}`, http.StatusBadRequest},
			{http.MethodGet, "/api/tasks/missing", "", http.StatusNotFound},
			{http.MethodPost, "/api/tasks/missing/complete", "", http.StatusNotFound},
			{http.MethodDelete, "/api/tasks/missing", "", http.StatusNotFound},
		} {
			response := app.HTTP(tc.method, tc.target, tc.body)
			assert.Equal(t, tc.status, response.Code, tc.target)
			assert.Contains(t, response.Body.String(), `"error"`, tc.target)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		response := demotest.Start(t).HTTP(http.MethodGet, "/api/tasks", "")
		require.Equal(t, http.StatusOK, response.Code)
		assert.JSONEq(t, `[]`, response.Body.String())
	})
}
