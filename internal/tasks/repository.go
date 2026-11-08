package tasks

import (
	"context"
	"database/sql"
	"errors"
)

// Store stores the tasks. The [Repository] implements it on the database; tests can replace it
// with fx.Replace, to test the service and the APIs without database, or with failures.
type Store interface {
	List(ctx context.Context) ([]Task, error)
	Get(ctx context.Context, id string) (Task, error)
	Insert(ctx context.Context, task Task) error
	Complete(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

// Repository is the [Store] on the database. Its queries are traced by the sql module.
// The placeholders ($1...) work with both PostgreSQL and SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository returns a [Repository]. Its dependencies are injected by fx.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
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

// Insert stores a new task.
func (r *Repository) Insert(ctx context.Context, task Task) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO tasks (id, title, done, created_at) VALUES ($1, $2, $3, $4)",
		task.ID, task.Title, task.Done, task.CreatedAt)

	return err
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
