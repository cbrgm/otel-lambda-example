package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	service           = "otel-lambda-test"
	environment       = "dev"
	id          int64 = 12345
	res               = resource.NewWithAttributes(
		semconv.SchemaURL,
		// sets the service correctly
		semconv.ServiceNameKey.String(service),

		// helps parse stack traces and errors
		semconv.TelemetrySDKLanguageGo,

		// others
		semconv.DeploymentEnvironmentKey.String(environment),
		attribute.Int64("id", id),
	)
)

func main() {
	ctx := context.Background()

	// Setup instrumentation
	mp := setupMetrics(ctx, res)
	tp := setupTracing(ctx, res)

	lambdaHandler := &LambdaHandler{
		metrics: mp,
	}

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	defer func() {
		if err := mp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()

	// Setup Lambda
	lambda.StartWithOptions(
		otellambda.InstrumentHandler(
			lambdaHandler.HandleRequest,
			xrayconfig.WithRecommendedOptions(tp)...,
		),
	)
}

type LambdaHandler struct {
	metrics metric.MeterProvider
}

func (h *LambdaHandler) HandleRequest(ctx context.Context, _ any) (any, error) {
	for i := 0; i < 10; i++ {
		doSomething(ctx, i, h.metrics)
	}
	return nil, nil
}

func doSomething(ctx context.Context, iteration int, mp metric.MeterProvider) {
	meter := mp.Meter("foo")
	ctx, span := otel.Tracer(service).Start(ctx, "proccessing")
	defer span.End()

	fmt.Printf("iteration: %d", iteration)

	if counter, Merr := meter.Int64Counter("requests"); Merr == nil {
		endpointName := "Endpoint"
		attrs := attribute.String("endpoint", endpointName)
		counter.Add(ctx, 1, attrs)
	}

	<-time.After(3 * time.Millisecond)
}
