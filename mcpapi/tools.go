// Package mcpapi exposes the tasks to AI agents: MCP tools on the [tasks.Service].
package mcpapi

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"

	"github.com/go-oryn/oryn/mcpserver"

	"github.com/go-oryn/demo/tasks"
)

// Module registers the tools on the MCP server.
var Module = fx.Module("mcpapi",
	mcpserver.AsCapability(NewTools),
)

// Tools are the MCP tools of the tasks.
type Tools struct {
	service *tasks.Service
}

// NewTools returns the [Tools]. Their dependencies are injected by fx.
func NewTools(service *tasks.Service) *Tools {
	return &Tools{service: service}
}

// TaskList is the output of the list_tasks tool.
type TaskList struct {
	Tasks []tasks.Task `json:"tasks"`
}

// CreateInput is the input of the create_task tool.
type CreateInput struct {
	Title string `json:"title" jsonschema:"the title of the task"`
}

// IDInput is the input of the tools acting on one task.
type IDInput struct {
	ID string `json:"id" jsonschema:"the ID of the task"`
}

// Register adds the tools to the MCP server. Their input and output JSON schemas are
// inferred from the Go types. An error returned by a tool is sent to the agent as a tool
// error, that the mcpserver module traces with error.type=tool_error.
func (t *Tools) Register(server *mcp.Server) error {
	mcp.AddTool(server, &mcp.Tool{Name: "list_tasks", Description: "List all the tasks, oldest first."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, TaskList, error) {
			list, err := t.service.List(ctx)

			return nil, TaskList{Tasks: list}, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "create_task", Description: "Create a task."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input CreateInput) (*mcp.CallToolResult, tasks.Task, error) {
			task, err := t.service.Create(ctx, input.Title)

			return nil, task, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "complete_task", Description: "Mark a task as done."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input IDInput) (*mcp.CallToolResult, tasks.Task, error) {
			task, err := t.service.Complete(ctx, input.ID)

			return nil, task, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "delete_task", Description: "Delete a task."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input IDInput) (*mcp.CallToolResult, any, error) {
			return nil, nil, t.service.Delete(ctx, input.ID)
		})

	return nil
}
