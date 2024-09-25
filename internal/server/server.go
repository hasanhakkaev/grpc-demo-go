package server

import (
	"context"
	"errors"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/hasanhakkaev/yqapp-demo/internal/domain"
	"github.com/hasanhakkaev/yqapp-demo/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Services groups all the services exposed by a single gRPC Server.
type Services struct {
	TaskService *service.TaskService
	Health      *health.Server
}

var serviceStatus = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "service_up",                                // Metric name
	Help: "Whether the service is up (1) or down (0)", // Metric description
})

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
	pprofServer   *http.Server
	taskChannel   chan *domain.Task
	taskLimiter   *rate.Limiter
}

// Run serves the application services.
func (s *Server) Run(ctx context.Context) error {
	go s.checkHealth(ctx)

	// Mark the service as up when starting
	markServiceUp()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.logger.Log(s.logger.Level(), "Starting Metrics endpoint /metrics", zap.String("port", s.cfg.GetConsumerMetricsPort()))
	go s.serveMetrics(ctx)

	s.logger.Log(s.logger.Level(), "Starting Pprof endpoint /debug/pprof", zap.String("port", s.cfg.GetConsumerProfilingPort()))
	go s.servePprof(ctx)

	s.logger.Log(s.logger.Level(), "Starting Consumer Service /debug/pprof", zap.Uint16("port", s.cfg.Server.Port))
	go s.startService(ctx)

	// Setup signal capturing for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigs:
		// Received shutdown signal (SIGINT or SIGTERM)
		s.logger.Log(s.logger.Level(), "Received shutdown signal", zap.String("signal", sig.String()))

		// Cancel the context to signal shutdown to goroutines
		cancel()
	}

	markServiceDown()

	err := s.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Shutdown releases any held resources by dependencies of this Server.
func (s *Server) Shutdown(ctx context.Context) error {
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
	}(s.db.DB, ctx)

	return err
}

func (s *Server) checkHealth(ctx context.Context) {
	s.logger.Log(s.logger.Level(), "Running health service")
	for {
		if ctx.Err() != nil {
			return
		}
		s.services.Health.SetServingStatus("app.db", s.checkDatabaseHealth(ctx))
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) checkDatabaseHealth(ctx context.Context) healthv1.HealthCheckResponse_ServingStatus {
	state := healthv1.HealthCheckResponse_SERVING
	err := s.db.DB.Ping(ctx)
	if err != nil {
		state = healthv1.HealthCheckResponse_NOT_SERVING
	}

	return state
}

func (s *Server) serveMetrics(ctx context.Context) {
	go func() {
		s.logger.Log(s.logger.Level(), "Metrics server started", zap.String("port", s.cfg.GetConsumerMetricsPort()))
		if err := s.metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to listen and serve metrics server", zap.Error(err))
		}
	}()

	// Wait for shutdown signals
	<-ctx.Done()

	// Gracefully shut down the metrics server
	s.logger.Log(s.logger.Level(), "Shutting down metrics server...")
	if err := s.metricsServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error during metrics server shutdown", zap.Error(err))
	}
}

func (s *Server) startService(ctx context.Context) {
	go func() {
		s.logger.Log(s.logger.Level(), "Running consumer service ", zap.String("port", strconv.Itoa(int(s.cfg.Server.Port))))
		if err := s.grpc.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to listen and serve consumer grpc server", zap.Error(err))
		}
	}()
	<-ctx.Done()

	// Gracefully shut down the metrics server
	s.logger.Log(s.logger.Level(), "Shutting down consumer service...")
	if err := s.metricsServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error during consumer service shutdown", zap.Error(err))
	}
}

func (s *Server) servePprof(ctx context.Context) {
	go func() {
		s.logger.Log(s.logger.Level(), "Pprof server started", zap.String("port", s.cfg.GetConsumerProfilingPort()))
		if err := s.pprofServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to listen and serve pprof server", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Gracefully shut down the metrics server
	s.logger.Log(s.logger.Level(), "Shutting down pprof server...")
	if err := s.pprofServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error during metrics pprof shutdown", zap.Error(err))
	}
}

func markServiceUp() {
	serviceStatus.Set(1) // Set to 1 when the service is up
}

func markServiceDown() {
	serviceStatus.Set(0) // Set to 0 when the service is down
}
