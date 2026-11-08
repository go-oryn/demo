// Package grpcapi (internal/grpc) is the gRPC API of the tasks: the TasksService of proto/tasks/v1/tasks.proto,
// on the [tasks.Service].
package grpcapi

import (
	"context"
	"errors"

	"go.uber.org/fx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-oryn/oryn/grpcserver"

	"github.com/go-oryn/demo/internal/tasks"
	tasksv1 "github.com/go-oryn/demo/proto/tasks/v1"
)

// Module registers the TasksService on the gRPC server.
var Module = fx.Module("grpcapi",
	grpcserver.AsService(&tasksv1.TasksService_ServiceDesc, NewTasksService),
)

// TasksService implements the gRPC TasksService.
type TasksService struct {
	tasksv1.UnimplementedTasksServiceServer

	service *tasks.Service
}

// NewTasksService returns a [TasksService]. Its dependencies are injected by fx.
func NewTasksService(service *tasks.Service) *TasksService {
	return &TasksService{service: service}
}

// ListTasks returns all the tasks, oldest first.
func (s *TasksService) ListTasks(ctx context.Context, _ *tasksv1.ListTasksRequest) (*tasksv1.ListTasksResponse, error) {
	list, err := s.service.List(ctx)
	if err != nil {
		return nil, grpcError(err)
	}

	response := &tasksv1.ListTasksResponse{Tasks: make([]*tasksv1.Task, 0, len(list))}
	for _, task := range list {
		response.Tasks = append(response.Tasks, toProto(task))
	}

	return response, nil
}

// GetTask returns a task, or NOT_FOUND.
func (s *TasksService) GetTask(ctx context.Context, request *tasksv1.GetTaskRequest) (*tasksv1.Task, error) {
	task, err := s.service.Get(ctx, request.GetId())
	if err != nil {
		return nil, grpcError(err)
	}

	return toProto(task), nil
}

// CreateTask creates a task, or returns INVALID_ARGUMENT.
func (s *TasksService) CreateTask(ctx context.Context, request *tasksv1.CreateTaskRequest) (*tasksv1.Task, error) {
	task, err := s.service.Create(ctx, request.GetTitle())
	if err != nil {
		return nil, grpcError(err)
	}

	return toProto(task), nil
}

// CompleteTask marks a task as done, or returns NOT_FOUND.
func (s *TasksService) CompleteTask(ctx context.Context, request *tasksv1.CompleteTaskRequest) (*tasksv1.Task, error) {
	task, err := s.service.Complete(ctx, request.GetId())
	if err != nil {
		return nil, grpcError(err)
	}

	return toProto(task), nil
}

// DeleteTask deletes a task, or returns NOT_FOUND.
func (s *TasksService) DeleteTask(ctx context.Context, request *tasksv1.DeleteTaskRequest) (*tasksv1.DeleteTaskResponse, error) {
	if err := s.service.Delete(ctx, request.GetId()); err != nil {
		return nil, grpcError(err)
	}

	return &tasksv1.DeleteTaskResponse{}, nil
}

// grpcError maps the tasks errors to gRPC status codes. The grpcserver module logs and traces them.
func grpcError(err error) error {
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, tasks.ErrInvalidTitle):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toProto(task tasks.Task) *tasksv1.Task {
	return &tasksv1.Task{
		Id:        task.ID,
		Title:     task.Title,
		Done:      task.Done,
		CreatedAt: timestamppb.New(task.CreatedAt),
	}
}
