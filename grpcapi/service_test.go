package grpcapi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-oryn/oryn/core/otel/oteltest/spantest"

	"github.com/go-oryn/demo/demotest"
	tasksv1 "github.com/go-oryn/demo/proto/tasks/v1"
)

func TestTasksService(t *testing.T) {
	t.Parallel()

	t.Run("task lifecycle", func(t *testing.T) {
		t.Parallel()

		app := demotest.Start(t)
		ctx := t.Context()

		task, err := app.GRPC.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "test_title"})
		require.NoError(t, err)
		assert.Equal(t, "test_title", task.GetTitle())

		completed, err := app.GRPC.CompleteTask(ctx, &tasksv1.CompleteTaskRequest{Id: task.GetId()})
		require.NoError(t, err)
		assert.True(t, completed.GetDone())

		list, err := app.GRPC.ListTasks(ctx, &tasksv1.ListTasksRequest{})
		require.NoError(t, err)
		require.Len(t, list.GetTasks(), 1)
		assert.Equal(t, task.GetId(), list.GetTasks()[0].GetId())

		_, err = app.GRPC.DeleteTask(ctx, &tasksv1.DeleteTaskRequest{Id: task.GetId()})
		require.NoError(t, err)

		_, err = app.GRPC.GetTask(ctx, &tasksv1.GetTaskRequest{Id: task.GetId()})
		assert.Equal(t, codes.NotFound, status.Code(err))

		// The tasks.create span is a child of the gRPC server span.
		server := spantest.Find(app.Spans, spantest.WithName("tasks.v1.TasksService/CreateTask"))
		require.Len(t, server, 1)
		spantest.HasSpan(t, app.Spans, spantest.WithName("tasks.create"), spantest.WithParent(server[0].SpanContext.SpanID()))
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		app := demotest.Start(t)
		ctx := t.Context()

		_, err := app.GRPC.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: " "})
		assert.Equal(t, codes.InvalidArgument, status.Code(err))

		_, err = app.GRPC.CompleteTask(ctx, &tasksv1.CompleteTaskRequest{Id: "missing"})
		assert.Equal(t, codes.NotFound, status.Code(err))

		_, err = app.GRPC.DeleteTask(ctx, &tasksv1.DeleteTaskRequest{Id: "missing"})
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}
