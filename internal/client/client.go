package client

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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"io"
	"net/http"
	"time"
)

// Services groups all the services exposed by a single gRPC Client.
type Services struct {
	TaskService v1.TaskServiceClient
	Health      *health.Server
}

// shutDowner holds a method to gracefully shut down a service or integration.
type shutDowner interface {
	// Shutdown releases any held computational resources.
	Shutdown(ctx context.Context) error
}

// Client abstracts all the functional components to be run by the server.
type Client struct {
	client        grpc.ClientConn
	listener      http.Handler
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
func (s *Client) Run(ctx context.Context) error {
	go s.checkHealth(ctx)
	go s.serveMetrics()

	client, err := grpc.NewClient(s.cfg.Server.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	s.logger.Info("Running Client")

	return nil
}

// Shutdown releases any held resources by dependencies of this Client.
func (s *Client) Shutdown(ctx context.Context) error {
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

func (s *Client) checkHealth(ctx context.Context) {
	s.logger.Info("Running health service")
	for {
		if ctx.Err() != nil {
			return
		}
		s.services.Health.SetServingStatus("app.db", s.checkDatabaseHealth())
		time.Sleep(10 * time.Second)
	}
}

func (s *Client) checkDatabaseHealth() healthv1.HealthCheckResponse_ServingStatus {
	state := healthv1.HealthCheckResponse_SERVING
	err := s.db.DB.Ping(context.Background())
	if err != nil {
		state = healthv1.HealthCheckResponse_NOT_SERVING
	}

	return state
}

func (s *Client) serveMetrics() {
	if err := s.metricsServer.ListenAndServe(); err != nil {
		s.logger.Error("failed to listen and server to metrics server", zap.Error(err))
	}
}
