package client

import (
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/domain"
	"github.com/hasanhakkaev/yqapp-demo/internal/telemetry"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net/http"
	_ "net/http/pprof" // Import pprof
)

// Setup creates a new client using the given ClientConfig.
func Setup(cfg conf.Configuration) (Client, error) {

	telemeter, err := telemetry.SetupTelemetry(cfg, "producer")
	if err != nil {
		return Client{}, err
	}

	telemeter.Logger.Debug("Initializing client", zap.String("client.name", cfg.Client.Name), zap.String("client.environment", cfg.Server.Environment))

	cc, err := grpc.NewClient(cfg.Server.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	taskClient := NewTaskClient(cc)

	prometheus.MustRegister(serviceStatus)

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.GetProducerMetricsPort()),
		Handler: promhttp.Handler(),
	}

	pprofServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.GetProducerProfilingPort()),
		Handler: http.DefaultServeMux,
	}

	limiter := rate.NewLimiter(rate.Limit(cfg.ProducerService.MessageProductionRate), 1)

	taskQueue := make(chan *domain.Task, cfg.ProducerService.MaxBacklog)

	return Client{
		task:          *taskClient,
		logger:        telemeter.Logger,
		meterProvider: telemeter.MeterProvider,
		shutdown: []shutDowner{
			telemeter.TraceExporter,
			telemeter.MeterExporter,
		},
		closer: []io.Closer{
			metricsServer,
		},
		cfg:           cfg,
		metricsServer: metricsServer,
		taskQueue:     taskQueue,
		pprofServer:   pprofServer,
		rateLimiter:   limiter,
	}, nil
}
