package telemetry

import (
	"github.com/hasanhakkaev/yqapp-demo/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// SetupLogger initializes a new Zap Logger with the parameters specified by the given ServerConfig.
func SetupLogger(logs conf.Logger) (*zap.Logger, error) {
	var (
		encoderCfg    zapcore.EncoderConfig
		isDevelopment bool
	)

	if !logs.Enabled {
		return zap.NewNop(), nil
	}

	level, err := zap.ParseAtomicLevel(logs.Level)
	if err != nil {
		return nil, err
	}

	switch logs.Environment {
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
		Encoding:          logs.Encoding,
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}

	return zap.Must(config.Build()), nil

}
