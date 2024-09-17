package telemetry

import (
	"github.com/hasanhakkaev/yqapp-demo/internal/config"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Telemetry struct {
	Logger         *zap.Logger
	TracerProvider trace.TracerProvider
	TraceExporter  tracesdk.SpanExporter
	MeterProvider  metric.MeterProvider
	MeterExporter  metricsdk.Exporter
	Propagator     propagation.TextMapPropagator
}

func SetupTelemetry(logging conf.Logger, tracing conf.Tracing, metrics conf.Metrics) (Telemetry, error) {
	var t Telemetry
	var err error

	t.Logger, err = SetupLogger(logging)
	if err != nil {
		return Telemetry{}, err
	}

	t.TracerProvider, t.TraceExporter, err = SetupTracing(tracing)
	if err != nil {
		return Telemetry{}, err
	}

	t.MeterProvider, t.MeterExporter, err = SetupMetrics(metrics)
	if err != nil {
		return Telemetry{}, err
	}

	t.Propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return t, nil
}
