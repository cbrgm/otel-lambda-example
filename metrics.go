package main

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
)

func setupMetrics(ctx context.Context, res *resource.Resource) {
	// communicate on localhost to the ADOT collector
	metricCollector, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("0.0.0.0:4317"),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		panic(err)
	}

	metricsProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricCollector)),
		metric.WithResource(res),
	)
	defer func() {
		if err = metricsProvider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	global.SetMeterProvider(metricsProvider)
}
