package server

import (
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/hasanhakkaev/yqapp-demo/internal/interceptors"
	"github.com/hasanhakkaev/yqapp-demo/internal/service"
	"github.com/hasanhakkaev/yqapp-demo/internal/telemetry"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"io"
	"net"
	"net/http"
)

func registerServices(srv *grpc.Server, svc Services) {
	v1.RegisterTaskServiceServer(srv, svc.TaskService)
	healthv1.RegisterHealthServer(srv, svc.Health)
}

// setupServices initializes the Server Services.
func setupServices(queries *database.Queries, logger *zap.Logger, meterProvider metric.MeterProvider) Services {
	logger.Debug("Initializing services")
	taskService := service.NewTaskService(logger, queries, meterProvider.Meter("task.service"))
	healthService := health.NewServer()
	return Services{
		TaskService: taskService,
		Health:      healthService,
	}
}

// setupListener initializes a new tcp listener used by a gRPC server.
func setupListener(cfg conf.Configuration, logger *zap.Logger) (net.Listener, error) {
	protocol, address := cfg.Server.Address()
	logger.Debug("Initializing listener", zap.String("listener.protocol", protocol), zap.String("listener.address", address))
	l, err := net.Listen(protocol, address)
	if err != nil {
		logger.Error("Failed to initialize listener", zap.Error(err))
		return nil, err
	}
	return l, nil
}

// setupDB initializes a new connection with a DB server.
func setupDB(cfg conf.Configuration, logger *zap.Logger) (*database.Postgres, error) {
	logger.Debug("Initializing DB connection", zap.String("db.engine", cfg.Database.Engine), zap.String("db.dsn", NewDSNFromConfig(cfg.Database)))

	db, err := database.NewPostgres(NewDSNFromConfig(cfg.Database))
	if err != nil {
		logger.Error("Failed to initialize DB connection", zap.Error(err))
		return nil, err
	}
	err = database.MigrateModels(NewDSNFromConfig(cfg.Database))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewDSNFromConfig(db conf.Database) string {
	return fmt.Sprintf("%s:%s@%s:%d/%s", db.Username, db.Password, db.Host, db.Port, db.Database)

}

// Setup creates a new application using the given ServerConfig.
func Setup(cfg conf.Configuration) (Server, error) {

	telemeter, err := telemetry.SetupTelemetry(cfg.Logger, cfg.Metrics)
	if err != nil {
		return Server{}, err
	}

	telemeter.Logger.Debug("Initializing server", zap.String("server.name", cfg.Server.Name), zap.String("server.environment", cfg.Server.Environment))

	db, err := setupDB(cfg, telemeter.Logger)
	if err != nil {
		return Server{}, err
	}

	queries := database.New(db.DB)

	l, err := setupListener(cfg, telemeter.Logger)
	if err != nil {
		return Server{}, err
	}

	srv := grpc.NewServer(interceptors.NewServerInterceptors(telemeter)...)
	reflection.Register(srv)

	svc := setupServices(queries, telemeter.Logger, telemeter.MeterProvider)
	registerServices(srv, svc)

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
		Handler: promhttp.Handler(),
	}

	return Server{
		grpc:          srv,
		listener:      l,
		logger:        telemeter.Logger,
		meterProvider: telemeter.MeterProvider,
		db:            db,
		services:      svc,
		metricsServer: metricsServer,
		shutdown: []shutDowner{
			telemeter.TraceExporter,
			telemeter.MeterExporter,
		},
		closer: []io.Closer{
			metricsServer,
		},
		cfg: cfg,
	}, nil
}
