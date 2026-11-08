// Package tasks is the tasks domain: a [Service] holding the business rules, on a [Repository]
// storing the tasks in the database. The HTTP, gRPC and MCP APIs all use the same [Service].
package tasks

import (
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a task does not exist.
	ErrNotFound = errors.New("task not found")
	// ErrInvalidTitle is returned when a task title is empty.
	ErrInvalidTitle = errors.New("the task title is required")
)

// Task is a thing to do.
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}
