package tasks_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/fx"

	"github.com/go-oryn/oryn/core/otel/oteltest/logtest"
	"github.com/go-oryn/oryn/core/otel/oteltest/metrictest"
	"github.com/go-oryn/oryn/core/otel/oteltest/spantest"

	"github.com/go-oryn/demo/internal/demotest"
	"github.com/go-oryn/demo/internal/tasks"
)

func TestService(t *testing.T) {
	t.Parallel()

	t.Run("task lifecycle", func(t *testing.T) {
		t.Parallel()

		var service *tasks.Service

		app := demotest.Start(t, fx.Populate(&service))

		task, err := service.Create(t.Context(), "  test_title  ")
		require.NoError(t, err)
		assert.Equal(t, "test_title", task.Title)
		assert.False(t, task.Done)

		completed, err := service.Complete(t.Context(), task.ID)
		require.NoError(t, err)
		assert.True(t, completed.Done)

		list, err := service.List(t.Context())
		require.NoError(t, err)
		assert.Equal(t, []tasks.Task{completed}, list)

		require.NoError(t, service.Delete(t.Context(), task.ID))

		_, err = service.Get(t.Context(), task.ID)
		require.ErrorIs(t, err, tasks.ErrNotFound)

		// A span per operation, parent of its SQL queries.
		create := spantest.Find(app.Spans, spantest.WithName("tasks.create"), spantest.WithAttributeValue(tasks.TaskIDKey.String(task.ID)))
		require.Len(t, create, 1)
		spantest.HasSpan(t, app.Spans, spantest.WithName("INSERT"), spantest.WithParent(create[0].SpanContext.SpanID()))
		spantest.HasSpan(t, app.Spans, spantest.WithName("tasks.complete"))
		spantest.HasSpan(t, app.Spans, spantest.WithName("tasks.delete"))

		// A log per change, and the counters.
		for _, body := range []string{"task created", "task completed", "task deleted"} {
			logtest.HasRecord(t, app.Logs, logtest.WithBody(body), logtest.WithLevel(slog.LevelInfo),
				logtest.WithAttributeValue("task_id", task.ID))
		}

		assert.Equal(t, int64(1), metrictest.Int64Sum(t, app.Metrics, "tasks.created"))
		assert.Equal(t, int64(1), metrictest.Int64Sum(t, app.Metrics, "tasks.completed"))
	})

	t.Run("invalid title", func(t *testing.T) {
		t.Parallel()

		var service *tasks.Service

		app := demotest.Start(t, fx.Populate(&service))

		_, err := service.Create(t.Context(), "  ")
		require.ErrorIs(t, err, tasks.ErrInvalidTitle)

		// An expected error: recorded on the span as an attribute, not as a failure.
		spantest.HasSpan(t, app.Spans,
			spantest.WithName("tasks.create"),
			spantest.WithStatus(codes.Unset),
			spantest.WithAttributeValue(attribute.String("tasks.error", tasks.ErrInvalidTitle.Error())),
		)
		logtest.HasNoRecord(t, app.Logs, logtest.WithBody("task created"))
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		var service *tasks.Service

		demotest.Start(t, fx.Populate(&service))

		_, err := service.Get(t.Context(), "missing")
		require.ErrorIs(t, err, tasks.ErrNotFound)

		_, err = service.Complete(t.Context(), "missing")
		require.ErrorIs(t, err, tasks.ErrNotFound)

		require.ErrorIs(t, service.Delete(t.Context(), "missing"), tasks.ErrNotFound)
	})

	t.Run("store failure", func(t *testing.T) {
		t.Parallel()

		var service *tasks.Service

		app := demotest.Start(t, demotest.ReplaceStore(demotest.FailingStore{}), fx.Populate(&service))

		_, err := service.Create(t.Context(), "test_title")
		require.ErrorIs(t, err, demotest.ErrStore)

		// An unexpected error: the span fails, and nothing is counted or logged.
		spantest.HasSpan(t, app.Spans, spantest.WithName("tasks.create"), spantest.WithStatus(codes.Error))
		logtest.HasNoRecord(t, app.Logs, logtest.WithBody("task created"))
		assert.Equal(t, int64(0), metrictest.Int64Sum(t, app.Metrics, "tasks.created"))
	})
}
