package client

import (
	"context"
	"errors"
	v1 "github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/domain"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// shutDowner holds a method to gracefully shut down a service or integration.
type shutDowner interface {
	// Shutdown releases any held computational resources.
	Shutdown(ctx context.Context) error
}

var (
	serviceStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "service_up",                                // Metric name
		Help: "Whether the service is up (1) or down (0)", // Metric description
	})
	producerTasks = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tasks_produced_total",
		Help: "The total number of produced tasks",
	})
)

// Client abstracts all the functional components to be run by the server.
type Client struct {
	task          v1.TaskServiceClient
	logger        *zap.Logger
	meterProvider metric.MeterProvider
	shutdown      []shutDowner
	closer        []io.Closer
	cfg           conf.Configuration
	metricsServer *http.Server
	taskQueue     chan *domain.Task // This is the backlog (buffered channel)
	rateLimiter   *rate.Limiter
	pprofServer   *http.Server
}

// Run serves the application services.
func (c *Client) Run(ctx context.Context) error {

	// Mark the service as up when starting
	markServiceUp()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.logger.Log(c.logger.Level(), "Starting Metrics endpoint /metrics", zap.String("port", c.cfg.GetConsumerMetricsPort()))
	go c.serveMetrics(ctx)

	c.logger.Log(c.logger.Level(), "Starting Pprof endpoint /metrics", zap.String("port", c.cfg.GetConsumerProfilingPort()))
	go c.servePprof(ctx)

	c.logger.Log(c.logger.Level(), "Running Client")

	go c.StartSending(ctx)

	// Start producing tasks at the specified rate
	productionDone := make(chan error, 1)
	go func() {
		productionDone <- c.ProduceTasks(ctx, int(c.cfg.ProducerService.MaxBacklog))
	}()

	// Setup signal capturing for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//err := c.ProduceTasks(context.Background(), int(c.cfg.ProducerService.MaxBacklog))
	//if err != nil {
	//	log.Fatalf("Error producing tasks: %v", err)
	//}
	//
	//<-sigs
	//c.logger.Log(c.logger.Level(), "Received shutdown signal")
	//
	//cancel()
	select {
	case sig := <-sigs:
		// Received shutdown signal (SIGINT or SIGTERM)
		c.logger.Log(c.logger.Level(), "Received shutdown signal", zap.String("signal", sig.String()))

		// Cancel the context to signal shutdown to goroutines
		cancel()

		// Wait for task production to finish
		err := <-productionDone
		if err != nil {
			c.logger.Error("Error during task production", zap.Error(err))
		}

	case err := <-productionDone:
		// Task production finished (all tasks produced)
		if err != nil {
			c.logger.Error("Error during task production", zap.Error(err))
		}
	}

	markServiceDown()

	err := c.Shutdown(context.Background())
	if err != nil {
		return err
	}

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

	return err
}

func (c *Client) serveMetrics(ctx context.Context) {
	//if err := c.metricsServer.ListenAndServe(); err != nil {
	//	c.logger.Error("failed to listen and server to metrics server", zap.Error(err))
	//}

	// Start the server in a separate goroutine
	go func() {
		c.logger.Log(c.logger.Level(), "Metrics server started", zap.String("port", c.cfg.GetProducerMetricsPort()))
		if err := c.metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.Error("failed to listen and serve metrics server", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Gracefully shut down the metrics server
	c.logger.Log(c.logger.Level(), "Shutting down metrics server...")
	if err := c.metricsServer.Shutdown(context.Background()); err != nil {
		c.logger.Error("Error during metrics server shutdown", zap.Error(err))
	}
}

func (c *Client) servePprof(ctx context.Context) {
	// Start the HTTP server for pprof on a specific port (e.g., 6060)
	//if err := c.pprofServer.ListenAndServe(); err != nil {
	//	log.Println("pprof server failed:", err)
	//}
	go func() {
		c.logger.Log(c.logger.Level(), "Pprof server started", zap.String("port", c.cfg.GetProducerProfilingPort()))
		if err := c.pprofServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.Error("failed to listen and serve pprof server", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Gracefully shut down the metrics server
	c.logger.Log(c.logger.Level(), "Shutting down pprof server...")
	if err := c.pprofServer.Shutdown(context.Background()); err != nil {
		c.logger.Error("Error during metrics pprof shutdown", zap.Error(err))
	}
}

// NewTaskClient returns a new task client
func NewTaskClient(cc *grpc.ClientConn) *v1.TaskServiceClient {
	client := v1.NewTaskServiceClient(cc)
	return &client
}

// ProduceTasks produces tasks at a controlled rate and enqueues them into the taskQueue.
func (c *Client) ProduceTasks(ctx context.Context, totalMessages int) error {
	for i := 0; i < totalMessages; i++ {
		// Wait until the rate limiter allows the next message to be produced
		if err := c.rateLimiter.Wait(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				c.logger.Warn("Task production stopped due to context cancellation")
				return nil // Gracefully stop without logging an error
			}

			c.logger.Error("Rate limiter error", zap.Error(err))
			return err
		}

		// Generate a new random task
		task := domain.RandomTask()

		// Attempt to enqueue the task into the backlog (taskQueue)
		select {
		case c.taskQueue <- task:
			c.logger.Log(c.logger.Level(), "Task produced and enqueued", zap.Int("task number", i+1))
		case <-ctx.Done():
			c.logger.Warn("Context cancelled, stopping task production")
			return ctx.Err()
		}
	}

	return nil
}

// StartSending starts sending tasks from the taskQueue to the server.
func (c *Client) StartSending(ctx context.Context) {
	for {
		select {
		case task := <-c.taskQueue:
			// Send the task to the server
			err := c.sendTask(ctx, task)
			if err != nil {
				c.logger.Error("Failed to send task", zap.Error(err))
			} else {
				c.logger.Log(c.logger.Level(), "Task sent successfully")
			}
		case <-ctx.Done():
			// Context canceled, initiate graceful shutdown
			c.logger.Warn("Context cancelled, stopping task sending. Draining remaining tasks...")

			// Drain the task queue to process remaining tasks
			for len(c.taskQueue) > 0 {
				task := <-c.taskQueue
				err := c.sendTask(ctx, task)
				if err != nil {
					c.logger.Error("Failed to send task during draining", zap.Error(err))
				} else {
					c.logger.Log(c.logger.Level(), "Task sent successfully during draining")
				}
			}

			c.logger.Warn("All remaining tasks processed, task sending stopped")
			return
		}
	}
}

// sendTask sends a task to the server using gRPC
func (c *Client) sendTask(ctx context.Context, task *domain.Task) error {
	// Convert domain task to protobuf task
	protoTask := domain.FromDomainToProto(task)

	// Create the gRPC request
	req := &v1.CreateTaskRequest{
		Task: protoTask,
	}

	// Call the gRPC CreateTask method
	_, err := c.task.CreateTask(ctx, req)

	producerTasks.Inc()

	return err
}

func markServiceUp() {
	serviceStatus.Set(1) // Set to 1 when the service is up
}

func markServiceDown() {
	serviceStatus.Set(0) // Set to 0 when the service is down
}
