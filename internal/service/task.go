package service

import (
	"context"
	"github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

type tasks struct {
	v1.UnimplementedTaskConsumerServiceServer
	v1.UnimplementedTaskProducerServiceServer
	logger *zap.Logger
	DB     *database.Postgres
	meter  metric.Meter
}

// NewTaskProducer initializes a new v1.TaskProducerServiceServer implementation.
func NewTaskProducer(logger *zap.Logger, postgres *database.Postgres, meter metric.Meter) *v1.TaskProducerServiceServer {
	return &tasks{
		logger: logger,
		DB:     postgres,
		meter:meter
	}
}

// NewTasksConsumer initializes a new v1.TaskConsumerServiceServer implementation.
func NewTasksConsumer(logger *zap.Logger, postgres *database.Postgres, meter metric.Meter) *v1.TaskConsumerServiceServer {
	tasksLogger := logger.Named("service.tasks.reader")
	return &tasks{
		DB:     postgres,
		logger: tasksLogger,
		meter:  meter,
	}
}

func (svc *tasks) CreateTask(ctx context.Context, request *v1.CreateTaskRequest) (*v1.Task, error) {
	svc.logger.Debug("Creating task", zap.String("task.title", strconv.Itoa(int(request.GetTask().GetId()))))
	span := trace.SpanFromContext(ctx)
	defer span.End()

	var task domain.Task
	svc.logger.Debug("Filling out task information")
	span.AddEvent("Parsing task from API request")
	task.FromAPI(request.GetTask())
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	span.AddEvent("Persisting task in the database")
	svc.logger.Debug("Persisting task in the database", zap.String("task.title", request.GetTask().GetTitle()))
	err := svc.db.Model(&domain.Task{}).WithContext(ctx).Create(&task).Error
	if err != nil {
		svc.logger.Error("Failed to create task", zap.Error(err))
		span.RecordError(err)
		return nil, status.Error(codes.Unavailable, "failed to create task")
	}
	svc.logger.Debug("Returning created task", zap.String("task.title", request.GetTask().GetTitle()))
	return task.API(), nil

}
