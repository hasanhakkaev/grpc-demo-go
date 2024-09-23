package client

import (
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/hasanhakkaev/yqapp-demo/internal/telemetry"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net/http"
)

// Setup creates a new application using the given ServerConfig.
func Setup(cfg conf.Configuration) (Client, error) {

	telemeter, err := telemetry.SetupTelemetry(cfg.Logger, cfg.Metrics)
	if err != nil {
		return Client{}, err
	}

	telemeter.Logger.Debug("Initializing client", zap.String("client.name", cfg.Server.Name), zap.String("server.environment", cfg.Server.Environment))

	db, err := setupDB(cfg, telemeter.Logger)
	if err != nil {
		return Client{}, err
	}

	queries := database.New(db.DB)

	cc, err := grpc.NewClient(cfg.Server.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	taskClient := NewTaskClient(cc)

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
		Handler: promhttp.Handler(),
	}

	return Client{
		task:          *taskClient,
		logger:        telemeter.Logger,
		db:            db,
		queries:       queries,
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
	}, nil
}

// setupDB initializes a new connection with a DB server.
func setupDB(cfg conf.Configuration, logger *zap.Logger) (*database.Postgres, error) {
	logger.Debug("Initializing DB connection", zap.String("db.engine", cfg.Database.Engine), zap.String("db.dsn", NewDSNFromConfig(cfg.Database)))

	db, err := database.NewPostgres(NewDSNFromConfig(cfg.Database))
	if err != nil {
		logger.Error("Failed to initialize DB connection", zap.Error(err))
		return nil, err
	}

	return db, nil
}

func NewDSNFromConfig(db conf.Database) string {
	return fmt.Sprintf("%s:%s@%s:%d/%s", db.Username, db.Password, db.Host, db.Port, db.Database)

}
