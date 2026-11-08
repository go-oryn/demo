package tasks

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-oryn/oryn/core/otel"
)

// Repository stores the tasks in the database. Its queries are traced by the sql module.
// The placeholders ($1...) work with both PostgreSQL and SQLite.
type Repository struct {
	db        *sql.DB
	telemetry otel.Telemetry
	created   metric.Int64Counter
}

// NewRepository returns a [Repository]. Its dependencies are injected by fx.
func NewRepository(db *sql.DB, telemetry otel.Telemetry) (*Repository, error) {
	created, err := telemetry.Meter().Int64Counter("tasks.created", metric.WithDescription("Number of created tasks"))
	if err != nil {
		return nil, err
	}

	return &Repository{db: db, telemetry: telemetry, created: created}, nil
}

// List returns all the tasks, oldest first.
func (r *Repository) List(ctx context.Context) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, done, created_at FROM tasks ORDER BY created_at, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// Get returns a task, or [ErrNotFound].
func (r *Repository) Get(ctx context.Context, id string) (Task, error) {
	var task Task

	err := r.db.QueryRowContext(ctx, "SELECT id, title, done, created_at FROM tasks WHERE id = $1", id).
		Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return task, ErrNotFound
	}

	return task, err
}

// Create stores a new task.
func (r *Repository) Create(ctx context.Context, title string) (Task, error) {
	task := Task{ID: newID(), Title: title, CreatedAt: time.Now().UTC().Truncate(time.Millisecond)}

	_, err := r.db.ExecContext(ctx, "INSERT INTO tasks (id, title, done, created_at) VALUES ($1, $2, $3, $4)",
		task.ID, task.Title, task.Done, task.CreatedAt)
	if err != nil {
		return task, err
	}

	// Enrich the request span, count the task, and log it: the log is correlated with the span.
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("task.id", task.ID))
	r.created.Add(ctx, 1)
	r.telemetry.Logger().InfoContext(ctx, "task created", "task_id", task.ID)

	return task, nil
}

// Complete marks a task as done, or returns [ErrNotFound].
func (r *Repository) Complete(ctx context.Context, id string) error {
	return r.exec(ctx, "UPDATE tasks SET done = TRUE WHERE id = $1", id)
}

// Delete deletes a task, or returns [ErrNotFound].
func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
}

// exec runs a statement on a task, and returns [ErrNotFound] if it did not exist.
func (r *Repository) exec(ctx context.Context, query string, id string) error {
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

// newID returns a random task ID.
func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("cannot generate a task id: %v", err))
	}

	return hex.EncodeToString(b)
}
