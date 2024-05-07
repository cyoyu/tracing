// Package tracing provides handy helper functions for measuring API performance and debugging.
package tracing

import (
	"context"
	"time"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

var tp *sdktrace.TracerProvider
var projectID string
var serviceName string
var deploymentEnvironment string
var tracerName string
var timeout int64

// Config provides customization of tracing
type Config struct {
	ProjectID             string `json:"project_id" yaml:"project_id"`
	TracerName            string `json:"tracer_name" yaml:"tracer_name"`
	ServiceName           string `json:"service_name" yaml:"service_name"`
	DeploymentEnvironment string `json:"deployment_environment" yaml:"deployment_environment"`
	TimeoutInSeconds      int64  `json:"timeout_in_seconds" yaml:"timeout_in_seconds"`
}

// Start creates a new span for a given name
func Start(ctx context.Context, name string) trace.Span {
	_, span := tp.Tracer(tracerName).Start(ctx, name)
	return span
}

// Initialize setup for tracing
func Initialize(ctx context.Context, c *Config) error {
	options := []texporter.Option{}
	if c != nil {
		projectID = c.ProjectID
		serviceName = c.ServiceName
		deploymentEnvironment = c.DeploymentEnvironment
		tracerName = c.TracerName
		timeout = c.TimeoutInSeconds
	}
	options = append(options, texporter.WithProjectID(projectID))
	if timeout > 0 {
		options = append(options, texporter.WithTimeout(time.Duration(timeout)*time.Second))
	}
	exporter, err := texporter.New(options...)
	if err != nil {
		return err
	}
	res, err := resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()),
		//resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.DeploymentEnvironmentKey.String(deploymentEnvironment),
		),
	)
	if err != nil {
		return err
	}

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return nil
}

// Finalize cleanup tracing
func Finalize(ctx context.Context) {
	if tp != nil {
		tp.Shutdown(ctx)
	}
}
