package client

import (
	"context"
	v1 "github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net/http"
	"time"
)

// Services groups all the services exposed by a single gRPC Client.
//type Services struct {
//	//TaskService v1.TaskServiceClient
//	Health *health.Server
//}

// shutDowner holds a method to gracefully shut down a service or integration.
type shutDowner interface {
	// Shutdown releases any held computational resources.
	Shutdown(ctx context.Context) error
}

// Client abstracts all the functional components to be run by the server.
type Client struct {
	task          v1.TaskServiceClient
	logger        *zap.Logger
	db            *database.Postgres
	queries       *database.Queries
	meterProvider metric.MeterProvider
	shutdown      []shutDowner
	closer        []io.Closer
	cfg           conf.Configuration
	metricsServer *http.Server
}

// Run serves the application services.
func (c *Client) Run(ctx context.Context) error {
	go c.serveMetrics()

	c.logger.Info("Running Client")

	return nil
}

// Shutdown releases any held resources by dependencies of this Client.
func (c *Client) Shutdown(ctx context.Context) error {
	var err error
	for _, downer := range c.shutdown {
		if downer == nil {
			continue
		}
		err = multierr.Append(err, downer.Shutdown(ctx))
	}
	for _, closer := range c.closer {
		if closer == nil {
			continue
		}
		err = multierr.Append(err, closer.Close())
	}

	defer func(db *pgxpool.Pool, ctx context.Context) {
		db.Close()
	}(c.db.DB, context.Background())

	return err
}

func (c *Client) checkDatabaseHealth() healthv1.HealthCheckResponse_ServingStatus {
	state := healthv1.HealthCheckResponse_SERVING
	err := c.db.DB.Ping(context.Background())
	if err != nil {
		state = healthv1.HealthCheckResponse_NOT_SERVING
	}

	return state
}

func (c *Client) serveMetrics() {
	if err := c.metricsServer.ListenAndServe(); err != nil {
		c.logger.Error("failed to listen and server to metrics server", zap.Error(err))
	}
}

// NewTaskClient returns a new task client
func NewTaskClient(cc *grpc.ClientConn) *v1.TaskServiceClient {
	client := v1.NewTaskServiceClient(cc)
	return &client
}

func (c *Client) CreateTask(task *v1.Task) {
	req := &v1.CreateTaskRequest{
		Task: task,
	}

	c.logger.Info("Creating task")

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.task.CreateTask(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			// not a big deal
			log.Print("laptop already exists")
		} else {
			log.Fatal("cannot create laptop: ", err)
		}
		return
	}

	log.Printf("created laptop with id: %s", res.Value)
}
