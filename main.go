package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/otel/attribute"
	metricsglobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
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

	setupMetrics(ctx, res)

	lambdaHandler := &LambdaHandler{}

	// Setup Lambda
	lambda.StartWithOptions(
		otellambda.InstrumentHandler(
			lambdaHandler.HandleRequest,
		),
	)
}

type LambdaHandler struct{}

func (h *LambdaHandler) HandleRequest(ctx context.Context, _ any) (any, error) {
	for i := 0; i < 10; i++ {
		doSomething(ctx, i)
	}
	return nil, nil
}

func doSomething(ctx context.Context, iteration int) {
	meter := metricsglobal.MeterProvider().Meter("foo")

	counter, err := meter.Int64Counter(
		"request_handled",
		instrument.WithUnit("1"),
		instrument.WithDescription("Requests Handled"),
	)
	if err != nil {
		fmt.Printf("counter failed: %s", err)
	}

	fmt.Printf("iteration: %d", iteration)

	counter.Add(ctx, 1)
	<-time.After(3 * time.Millisecond)
}
