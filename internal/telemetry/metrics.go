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

func SetupMetrics(conf conf.Configuration, targetService string) (metric.MeterProvider, metricsdk.Exporter, error) {
	if !conf.Metrics.Enabled {
		return noop.NewMeterProvider(), nil, nil
	}

	var meterProvider metric.MeterProvider
	var meterExporter metricsdk.Exporter

	switch conf.Metrics.Environment {
	case "production", "staging":
		var err error
		meterProvider, meterExporter, err = newMetrics(conf, targetService)
		if err != nil {
			return nil, nil, err
		}
	default:
		meterProvider = noop.NewMeterProvider()
	}

	return meterProvider, meterExporter, nil
}

func newMetrics(conf conf.Configuration, targetService string) (metric.MeterProvider, metricsdk.Exporter, error) {
	ctx := context.Background()
	res, err := newResource(ctx, "metrics", conf.Metrics.Environment)
	if err != nil {
		return nil, nil, err
	}
	var address string
	switch targetService {
	case "producer":
		address = fmt.Sprintf("%s:%s", "0.0.0.0", conf.GetProducerMetricsPort())

	default:
		address = fmt.Sprintf("%s:%s", "0.0.0.0", conf.GetConsumerMetricsPort())
	}
	conn, err := newClient(address)
	if err != nil {
		return nil, nil, err
	}
	meterExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}
	meterProvider := metricsdk.NewMeterProvider(
		metricsdk.WithReader(metricsdk.NewPeriodicReader(meterExporter)),
		metricsdk.WithResource(res),
	)
	return meterProvider, meterExporter, nil
}
