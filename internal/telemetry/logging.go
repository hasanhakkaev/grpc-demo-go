package telemetry

import (
	"github.com/hasanhakkaev/yqapp-demo/internal/config"
	"go.uber.org/zap"
)

// SetupLogger initializes a new Zap Logger with the parameters specified by the given ServerConfig.
func SetupLogger(logs conf.Logger) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error
	switch logs.Environment {
	case "production":
		logger, err = zap.NewProduction()
	case "staging":
		logger, err = zap.NewDevelopment()
	default:
		logger = zap.NewNop()
	}
	if err != nil {
		return nil, err
	}
	logger = logger.Named(logs.Name)
	return logger, nil
}
