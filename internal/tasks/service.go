package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-oryn/oryn/core/otel"
)

// TaskIDKey is the span attribute holding the task ID.
const TaskIDKey = attribute.Key("task.id")

// Service holds the tasks business rules, on a [Store]. It is used by the HTTP, gRPC and MCP APIs alike.
//
// It is instrumented with the application telemetry: a span per operation (tasks.create...),
// child of the API span and parent of the SQL query spans; the tasks.created and tasks.completed
// counters; and a log per change, correlated with the trace.
type Service struct {
	store     Store
	telemetry otel.Telemetry
	created   metric.Int64Counter
	completed metric.Int64Counter
}

// NewService returns a [Service]. Its dependencies are injected by fx.
func NewService(store Store, telemetry otel.Telemetry) (*Service, error) {
	meter := telemetry.Meter()

	created, err := meter.Int64Counter("tasks.created", metric.WithDescription("Number of created tasks"))
	if err != nil {
		return nil, err
	}

	completed, err := meter.Int64Counter("tasks.completed", metric.WithDescription("Number of completed tasks"))
	if err != nil {
		return nil, err
	}

	return &Service{store: store, telemetry: telemetry, created: created, completed: completed}, nil
}

// List returns all the tasks, oldest first.
func (s *Service) List(ctx context.Context) ([]Task, error) {
	ctx, span := s.start(ctx, "tasks.list")
	defer span.End()

	tasks, err := s.store.List(ctx)
	if err != nil {
		return nil, s.fail(span, err)
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))

	return tasks, nil
}

// Get returns a task, or [ErrNotFound].
func (s *Service) Get(ctx context.Context, id string) (Task, error) {
	ctx, span := s.start(ctx, "tasks.get", TaskIDKey.String(id))
	defer span.End()

	task, err := s.store.Get(ctx, id)
	if err != nil {
		return task, s.fail(span, err)
	}

	return task, nil
}

// Create creates a task, or returns [ErrInvalidTitle].
func (s *Service) Create(ctx context.Context, title string) (Task, error) {
	ctx, span := s.start(ctx, "tasks.create")
	defer span.End()

	task := Task{ID: newID(), Title: strings.TrimSpace(title), CreatedAt: time.Now().UTC().Truncate(time.Millisecond)}
	if task.Title == "" {
		return Task{}, s.fail(span, ErrInvalidTitle)
	}

	span.SetAttributes(TaskIDKey.String(task.ID))

	if err := s.store.Insert(ctx, task); err != nil {
		return Task{}, s.fail(span, err)
	}

	s.created.Add(ctx, 1)
	s.telemetry.Logger().InfoContext(ctx, "task created", "task_id", task.ID)

	return task, nil
}

// Complete marks a task as done, and returns it, or [ErrNotFound].
func (s *Service) Complete(ctx context.Context, id string) (Task, error) {
	ctx, span := s.start(ctx, "tasks.complete", TaskIDKey.String(id))
	defer span.End()

	if err := s.store.Complete(ctx, id); err != nil {
		return Task{}, s.fail(span, err)
	}

	task, err := s.store.Get(ctx, id)
	if err != nil {
		return Task{}, s.fail(span, err)
	}

	s.completed.Add(ctx, 1)
	s.telemetry.Logger().InfoContext(ctx, "task completed", "task_id", id)

	return task, nil
}

// Delete deletes a task, or returns [ErrNotFound].
func (s *Service) Delete(ctx context.Context, id string) error {
	ctx, span := s.start(ctx, "tasks.delete", TaskIDKey.String(id))
	defer span.End()

	if err := s.store.Delete(ctx, id); err != nil {
		return s.fail(span, err)
	}

	s.telemetry.Logger().InfoContext(ctx, "task deleted", "task_id", id)

	return nil
}

// start starts the span of an operation, with the application tracer.
func (s *Service) start(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	return s.telemetry.Tracer().Start(ctx, name, trace.WithAttributes(attributes...))
}

// fail records an error on the span: as an error status for unexpected errors, and as an
// attribute only for the expected ones (not found, invalid input), that are not failures of the service.
func (s *Service) fail(span trace.Span, err error) error {
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalidTitle) {
		span.SetAttributes(attribute.String("tasks.error", err.Error()))

		return err
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	return err
}

// newID returns a random task ID.
func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("cannot generate a task id: %v", err))
	}

	return hex.EncodeToString(b)
}
