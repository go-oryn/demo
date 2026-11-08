package mcpapi_test

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	"github.com/go-oryn/oryn/core/otel/oteltest/spantest"

	"github.com/go-oryn/demo/demotest"
	"github.com/go-oryn/demo/mcpapi"
	"github.com/go-oryn/demo/tasks"
)

// call calls a tool, and decodes its structured output into out.
func call(t *testing.T, app *demotest.App, tool string, arguments any, out any) *mcp.CallToolResult {
	t.Helper()

	result, err := app.MCP.CallTool(t.Context(), &mcp.CallToolParams{Name: tool, Arguments: arguments})
	require.NoError(t, err)

	if out != nil && !result.IsError {
		data, err := json.Marshal(result.StructuredContent)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, out))
	}

	return result
}

func TestTools(t *testing.T) {
	t.Parallel()

	t.Run("lists the tools", func(t *testing.T) {
		t.Parallel()

		result, err := demotest.Start(t).MCP.ListTools(t.Context(), nil)
		require.NoError(t, err)

		var names []string
		for _, tool := range result.Tools {
			names = append(names, tool.Name)
		}

		assert.ElementsMatch(t, []string{"list_tasks", "create_task", "complete_task", "delete_task"}, names)
	})

	t.Run("task lifecycle", func(t *testing.T) {
		t.Parallel()

		app := demotest.Start(t)

		var task tasks.Task
		require.False(t, call(t, app, "create_task", mcpapi.CreateInput{Title: "test_title"}, &task).IsError)
		assert.Equal(t, "test_title", task.Title)

		var completed tasks.Task
		require.False(t, call(t, app, "complete_task", mcpapi.IDInput{ID: task.ID}, &completed).IsError)
		assert.True(t, completed.Done)

		var list mcpapi.TaskList
		require.False(t, call(t, app, "list_tasks", struct{}{}, &list).IsError)
		require.Len(t, list.Tasks, 1)

		require.False(t, call(t, app, "delete_task", mcpapi.IDInput{ID: task.ID}, nil).IsError)

		// The tasks.create span is a child of the MCP tool call span.
		server := spantest.Find(app.Spans, spantest.WithName("tools/call create_task"))
		require.Len(t, server, 1)
		spantest.HasSpan(t, app.Spans, spantest.WithName("tasks.create"), spantest.WithParent(server[0].SpanContext.SpanID()))
	})

	t.Run("tool errors", func(t *testing.T) {
		t.Parallel()

		app := demotest.Start(t)

		assert.True(t, call(t, app, "create_task", mcpapi.CreateInput{Title: " "}, nil).IsError)
		assert.True(t, call(t, app, "complete_task", mcpapi.IDInput{ID: "missing"}, nil).IsError)

		spantest.HasSpan(t, app.Spans,
			spantest.WithName("tools/call complete_task"),
			spantest.WithAttributeValue(attribute.String("error.type", "tool_error")),
		)
	})
}
