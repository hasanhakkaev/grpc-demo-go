package interceptors

import (
	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/hasanhakkaev/yqapp-demo/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func NewServerInterceptors(telemeter telemetry.Telemetry) []grpc.ServerOption {
	var opts []grpc.ServerOption
	return append(opts,
		newServerUnaryInterceptors(telemeter),
		newServerStreamInterceptors(telemeter),
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(
				otelgrpc.WithTracerProvider(telemeter.TracerProvider),
				otelgrpc.WithMeterProvider(telemeter.MeterProvider),
				otelgrpc.WithPropagators(telemeter.Propagator),
			),
		),
	)
}

func newServerUnaryInterceptors(telemeter telemetry.Telemetry) grpc.ServerOption {
	var interceptors []grpc.UnaryServerInterceptor

	if telemeter.Logger != nil {
		interceptors = append(interceptors,
			grpclogging.UnaryServerInterceptor(interceptorLogger(telemeter.Logger)),
			grpcrecovery.UnaryServerInterceptor(grpcrecovery.WithRecoveryHandler(RecoveryHandler(telemeter.Logger))),
		)
	}

	return grpc.ChainUnaryInterceptor(interceptors...)
}

func newServerStreamInterceptors(telemeter telemetry.Telemetry) grpc.ServerOption {
	var interceptors []grpc.StreamServerInterceptor

	if telemeter.Logger != nil {
		interceptors = append(interceptors,
			grpclogging.StreamServerInterceptor(interceptorLogger(telemeter.Logger)),
			grpcrecovery.StreamServerInterceptor(grpcrecovery.WithRecoveryHandler(RecoveryHandler(telemeter.Logger))),
		)
	}

	return grpc.ChainStreamInterceptor(interceptors...)
}
