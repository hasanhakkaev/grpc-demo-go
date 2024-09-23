package client

import (
	"context"
	v1 "github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/domain"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
)

// shutDowner holds a method to gracefully shut down a service or integration.
type shutDowner interface {
	// Shutdown releases any held computational resources.
	Shutdown(ctx context.Context) error
}

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
}

// Run serves the application services.
func (c *Client) Run(ctx context.Context) error {
	go c.serveMetrics()

	c.logger.Info("Running Client")

	//ticker := time.NewTicker(time.Second / time.Duration(c.cfg.ProducerService.MessageProductionRate))
	//defer ticker.Stop() // Stop the ticker when we're done
	//
	//for i := 0; i < 1000; i++ {
	//	task := domain.RandomTask()
	//	err := c.CreateTask(*task)
	//	time.Sleep(0 * time.Second)
	//	if err != nil {
	//		return err
	//	}
	//	c.logger.Info("Task:", zap.Int("task", i))
	//	// Wait for the ticker before producing the next task
	//	<-ticker
	//
	//}
	go c.StartSending(context.Background())
	// Start producing tasks at the specified rate
	totalMessages := 1000 // Number of messages to produce
	err := c.ProduceTasks(context.Background(), totalMessages)
	if err != nil {
		log.Fatalf("Error producing tasks: %v", err)
	}

	err = c.Shutdown(context.Background())
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

//func (c *Client) CreateTask(dTask domain.Task) error {
//
//	c.logger.Info("Requesting task")
//
//	// set timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	protoTask := domain.FromDomainToProto(&dTask)
//
//	req := &v1.CreateTaskRequest{
//		Task: protoTask,
//	}
//
//	resp, err := c.task.CreateTask(ctx, req)
//	if err != nil {
//		st, ok := status.FromError(err)
//		if ok && st.Code() == codes.AlreadyExists {
//			log.Print("Task already exists")
//		} else {
//			log.Fatal("cannot create Task: ", err)
//		}
//		return nil
//	}
//
//	c.logger.Log(c.logger.Level(), "Returning created task", zap.String("task.id", strconv.Itoa(int(resp.Id))))
//
//	return nil
//
//}

// ProduceTasks produces tasks at a controlled rate and enqueues them into the taskQueue.
func (c *Client) ProduceTasks(ctx context.Context, totalMessages int) error {
	for i := 0; i < totalMessages; i++ {
		// Wait until the rate limiter allows the next message to be produced
		if err := c.rateLimiter.Wait(ctx); err != nil {
			c.logger.Error("Rate limiter error", zap.Error(err))
			return err
		}

		// Generate a new random task
		task := domain.RandomTask()

		// Attempt to enqueue the task into the backlog (taskQueue)
		select {
		case c.taskQueue <- task:
			c.logger.Info("Task produced and enqueued", zap.Int("task number", i+1))
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
				c.logger.Info("Task sent successfully", zap.Uint32("task_id", task.ID))
			}
		case <-ctx.Done():
			c.logger.Warn("Context cancelled, stopping task sending")
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
	return err
}
