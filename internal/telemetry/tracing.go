package telemetry

import (
	"context"
	"fmt"
	"github.com/hasanhakkaev/yqapp-demo/internal/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func SetupTracing(tracing conf.Tracing) (trace.TracerProvider, tracesdk.SpanExporter, error) {
	if !tracing.Enabled {
		return noop.NewTracerProvider(), nil, nil
	}

	var tracerProvider trace.TracerProvider
	var traceExporter tracesdk.SpanExporter
	switch tracing.Environment {
	case "production", "staging":
		var err error
		tracerProvider, traceExporter, err = newTracing(tracing)
		if err != nil {
			return nil, nil, err
		}
	default:
		tracerProvider = noop.NewTracerProvider()
	}

	return tracerProvider, traceExporter, nil
}

func newTracing(tracing conf.Tracing) (trace.TracerProvider, tracesdk.SpanExporter, error) {
	ctx := context.Background()
	res, err := newResource(ctx, "traces", tracing.Environment)
	if err != nil {
		return nil, nil, err
	}
	conn, err := newClient(tracing.Address())
	if err != nil {
		return nil, nil, err
	}
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	propagation.NewCompositeTextMapPropagator()
	bsp := tracesdk.NewBatchSpanProcessor(traceExporter)
	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithResource(res),
		tracesdk.WithSpanProcessor(bsp),
	)
	return tracerProvider, traceExporter, nil
}
