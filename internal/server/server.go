package server

import (
	"context"
	"github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"io"
	"net"
	"net/http"
	"time"
)

// Services groups all the services exposed by a single gRPC Server.
type Services struct {
	TaskService v1.TaskServiceServer
	Health      *health.Server
}

// shutDowner holds a method to gracefully shut down a service or integration.
type shutDowner interface {
	// Shutdown releases any held computational resources.
	Shutdown(ctx context.Context) error
}

// grpcServer holds the method to serve a gRPC server using a net.Listener.
type grpcServer interface {
	// Serve serves a gRPC server through net.Listener until an error occurs.
	Serve(net.Listener) error
}

// Server abstracts all the functional components to be run by the server.
type Server struct {
	grpc          grpcServer
	listener      net.Listener
	logger        *zap.Logger
	db            *database.Postgres
	services      Services
	meterProvider metric.MeterProvider
	shutdown      []shutDowner
	closer        []io.Closer
	cfg           conf.Configuration
	metricsServer *http.Server
}

// Run serves the application services.
func (s Server) Run(ctx context.Context) error {
	go s.checkHealth(ctx)
	go s.serveMetrics()

	s.logger.Info("Running Server")
	return s.grpc.Serve(s.listener)
}

// Shutdown releases any held resources by dependencies of this Server.
func (s Server) Shutdown(ctx context.Context) error {
	var err error
	for _, downer := range s.shutdown {
		if downer == nil {
			continue
		}
		err = multierr.Append(err, downer.Shutdown(ctx))
	}
	for _, closer := range s.closer {
		if closer == nil {
			continue
		}
		err = multierr.Append(err, closer.Close())
	}

	defer func(db *pgxpool.Pool, ctx context.Context) {
		db.Close()
	}(s.db.DB, context.Background())

	return err
}

func (s Server) checkHealth(ctx context.Context) {
	s.logger.Info("Running health service")
	for {
		if ctx.Err() != nil {
			return
		}
		s.services.Health.SetServingStatus("app.db", s.checkDatabaseHealth())
		time.Sleep(10 * time.Second)
	}
}

func (s Server) checkDatabaseHealth() healthv1.HealthCheckResponse_ServingStatus {
	state := healthv1.HealthCheckResponse_SERVING
	err := s.db.DB.Ping(context.Background())
	if err != nil {
		state = healthv1.HealthCheckResponse_NOT_SERVING
	}

	return state
}

func (s Server) serveMetrics() {
	if err := s.metricsServer.ListenAndServe(); err != nil {
		s.logger.Error("failed to listen and server to metrics server", zap.Error(err))
	}
}
