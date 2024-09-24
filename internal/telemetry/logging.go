package telemetry

import (
	"github.com/hasanhakkaev/yqapp-demo/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SetupLogger initializes a new Zap Logger with the parameters specified by the given ServerConfig.
func SetupLogger(cfg conf.Configuration, targetService string) (*zap.Logger, error) {
	var (
		encoderCfg    zapcore.EncoderConfig
		isDevelopment bool
		logLevel      string
		logEncoding   string
	)

	if !cfg.Logger.Enabled {
		return zap.NewNop(), nil
	}

	switch targetService {
	case "producer":
		logLevel = cfg.GetProducerLogLevel()
		logEncoding = cfg.GetProducerLogEncoding()
	default:
		logLevel = cfg.GetConsumerLogLevel()
		logEncoding = cfg.GetConsumerLogEncoding()
	}

	level, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		return nil, err
	}

	switch cfg.Logger.Environment {
	case "production":
		encoderCfg = zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		isDevelopment = false

	default:
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		isDevelopment = true
	}

	config := zap.Config{
		Level:             level,
		Development:       isDevelopment,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          logEncoding,
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	return zap.Must(config.Build()), nil

}
