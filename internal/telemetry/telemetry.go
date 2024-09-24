package telemetry

import (
	"github.com/hasanhakkaev/yqapp-demo/internal/config"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Telemetry struct {
	Logger         *zap.Logger
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	MeterExporter  metricsdk.Exporter
	Propagator     propagation.TextMapPropagator
}

func SetupTelemetry(cfg conf.Configuration, targetService string) (Telemetry, error) {
	var t Telemetry
	var err error

	t.Logger, err = SetupLogger(cfg, targetService)
	if err != nil {
		return Telemetry{}, err
	}

	t.MeterProvider, t.MeterExporter, err = SetupMetrics(cfg, targetService)
	if err != nil {
		return Telemetry{}, err
	}

	t.Propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return t, nil
}
