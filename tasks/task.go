// Package tasks is the tasks API: a repository on the database, and its HTTP handlers.
package tasks

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a task does not exist.
var ErrNotFound = errors.New("task not found")

// Task is a thing to do.
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}
