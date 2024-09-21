package telemetry

import (
	"context"
	"fmt"
	"github.com/hasanhakkaev/yqapp-demo/internal/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

func SetupMetrics(metrics conf.Metrics) (metric.MeterProvider, metricsdk.Exporter, error) {
	if !metrics.Enabled {
		return noop.NewMeterProvider(), nil, nil
	}

	var meterProvider metric.MeterProvider
	var meterExporter metricsdk.Exporter

	switch metrics.Environment {
	case "production", "staging":
		var err error
		meterProvider, meterExporter, err = newMetrics(metrics)
		if err != nil {
			return nil, nil, err
		}
	default:
		meterProvider = noop.NewMeterProvider()
	}

	return meterProvider, meterExporter, nil
}

func newMetrics(metrics conf.Metrics) (metric.MeterProvider, metricsdk.Exporter, error) {
	ctx := context.Background()
	res, err := newResource(ctx, "metrics", metrics.Environment)
	if err != nil {
		return nil, nil, err
	}
	conn, err := newClient(metrics.Address())
	if err != nil {
		return nil, nil, err
	}
	meterExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}
	meterProvider := metricsdk.NewMeterProvider(
		metricsdk.WithReader(metricsdk.NewPeriodicReader(meterExporter)),
		metricsdk.WithResource(res),
	)
	return meterProvider, meterExporter, nil
}
