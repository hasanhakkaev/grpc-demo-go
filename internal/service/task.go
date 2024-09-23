package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/hasanhakkaev/yqapp-demo/internal/domain"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

var (
	processingTasks = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tasks_processing_total",
		Help: "The total number of tasks being processed",
	})

	doneTasks = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tasks_done_total",
		Help: "The total number of tasks completed",
	})

	taskTypeCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tasks_per_type_total",
			Help: "The number of tasks per task type",
		},
		[]string{"type"},
	)

	taskValueSum = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_value_sum_per_type",
			Help: "The total sum of task values per task type",
		},
		[]string{"type"},
	)
)

type tasks struct {
	v1.UnimplementedTaskServiceServer
	logger  *zap.Logger
	queries *database.Queries
	meter   metric.Meter
}

// NewTaskService initializes a new v1.TaskProducerServiceServer implementation.
func NewTaskService(logger *zap.Logger, queries *database.Queries, meter metric.Meter) v1.TaskServiceServer {
	return &tasks{
		logger:  logger,
		queries: queries,
		meter:   meter,
	}
}

func (svc *tasks) CreateTask(ctx context.Context, request *v1.CreateTaskRequest) (*v1.Task, error) {

	svc.logger.Log(svc.logger.Level(), "Creating task")
	span := trace.SpanFromContext(ctx)
	defer span.End()

	var task domain.Task

	svc.logger.Log(svc.logger.Level(), "Filling out task information")
	span.AddEvent("Parsing task from API request")

	task.FromAPI(request.GetTask())

	task.CreationTime = float64(float32(time.Now().Unix()))
	task.LastUpdateTime = 0

	span.AddEvent("Persisting task in the database")

	taskFromDB, err := svc.queries.CreateTask(ctx, *task.ToTaskCreateParams())
	if err != nil {
		svc.logger.Log(svc.logger.Level(), "Failed to create task", zap.Error(err))
		span.RecordError(err)
		return nil, status.Error(codes.Unavailable, "failed to create task")
	}

	svc.logger.Log(svc.logger.Level(), "Returning created task", zap.String("task.id", strconv.Itoa(int(taskFromDB.ID))))
	return task.API(), nil

}

// ProcessTask processes a single task, updating its state and tracking metrics.
func (svc *tasks) ProcessTask(ctx context.Context, task *domain.Task) error {
	svc.logger.Info("Handling task", zap.Int("task.id", int(task.ID)))

	// Update task state to "processing"
	_, err := svc.queries.UpdateTaskState(ctx, database.UpdateTaskStateParams{
		State:          database.StatePROCESSING,
		LastUpdateTime: float64(time.Now().Unix()),
		ID:             int32(task.ID),
	})
	if err != nil {
		svc.logger.Error("Failed to update task to processing", zap.Error(err))
		return status.Error(codes.Internal, "Failed to update task to processing")
	}

	// Increment processing tasks metric
	processingTasks.Inc()

	// Simulate processing by sleeping for task's value in milliseconds
	time.Sleep(time.Duration(task.Value) * time.Millisecond)

	if errors.Is(ctx.Err(), context.Canceled) {
		svc.logger.Log(svc.logger.Level(), "Request is canceled")
		return status.Error(codes.Canceled, "Request is canceled")
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		svc.logger.Log(svc.logger.Level(), "Request deadline exceeded")
		return status.Error(codes.DeadlineExceeded, "Request deadline exceeded")
	}

	// Update task state to "done"
	_, err = svc.queries.UpdateTaskState(ctx, database.UpdateTaskStateParams{
		State:          database.StateDONE,
		LastUpdateTime: float64(time.Now().Unix()),
		ID:             int32(task.ID),
	})
	if err != nil {
		svc.logger.Error("Failed to update task to done", zap.Error(err))
		return status.Error(codes.Internal, "Failed to update task to done")
	}

	// Update metrics
	processingTasks.Dec()
	doneTasks.Inc()

	taskTypeCount.WithLabelValues(fmt.Sprintf("%d", task.Type)).Inc()
	taskValueSum.WithLabelValues(fmt.Sprintf("%d", task.Type)).Add(float64(task.Value))

	// Log the final task content and total sum for that type
	//totalSum := taskValueSum.WithLabelValues(fmt.Sprintf("%d", task.Type))

	svc.logger.Info("Task processed", zap.Int("id", int(task.ID)),
		zap.Int("type", int(task.Type)), zap.Int("value", int(task.Value)))
	return nil
}

// ConsumeTasks handles incoming tasks with a rate limiter.
func (svc *tasks) ConsumeTasks(taskChannel <-chan *domain.Task, limiter *rate.Limiter) {
	for task := range taskChannel {
		// Apply rate limiting
		err := limiter.Wait(context.Background())
		if err != nil {
			svc.logger.Fatal("Rate limiter error", zap.Error(err))
		}

		// Process each task
		err = svc.ProcessTask(context.Background(), task)
		if err != nil {
			return
		}
	}
}
