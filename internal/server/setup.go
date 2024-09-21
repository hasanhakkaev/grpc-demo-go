package server

import (
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/hasanhakkaev/yqapp-demo/internal/service"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func registerServices(srv *grpc.Server, svc Services) {
	ta.RegisterTasksWriterServiceServer(srv, svc.TasksWriter)
	tasksv1.RegisterTasksReaderServiceServer(srv, svc.TasksReader)
	healthv1.RegisterHealthServer(srv, svc.Health)
}

// setupServices initializes the Application Services.
func setupServices(postgres database.Postgres, logger *zap.Logger, tracerProvider trace.TracerProvider, meterProvider metric.MeterProvider) Services {
	logger.Debug("Initializing services")
	tasksWriterService := service.NewTaskProducer(logger, postgres, meterProvider.Meter("todo.huck.com.ar/tasks.writer"))
	tasksReaderService := service.NewTaskProducer(logger, postgres, meterProvider.Meter("todo.huck.com.ar/tasks.reader"))
	healthService := health.NewServer()
	return Services{
		Consumer: tasksWriterService,
		Producer: tasksReaderService,
		Health:   healthService,
	}
}
