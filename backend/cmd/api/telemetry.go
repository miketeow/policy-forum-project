package main

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func initTracer() *sdktrace.TracerProvider {
	// 1. Define the resource (app name)
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("policy-forum-backend"),
	)

	// 2. Check the environment variable
	enableTracing := os.Getenv("ENABLE_TRACING") == "true"

	if enableTracing {
		// ==========================================
		// LOCAL DEV: FULL TRACING WITH JAEGER
		// ==========================================
		exporter, err := otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpoint("localhost:4318"),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			log.Fatalf("failed to initialize exporter: %v", err)
		}

		// Include the Batcher (Exporter) in the provider
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)

		log.Println("OpenTelemetry initialized WITH Jaeger exporter")
		return tp
	}

	// ==========================================
	// PRODUCTION: LOG CORRELATION MODE ONLY
	// ==========================================
	// We create a provider WITHOUT an exporter (No WithBatcher).
	// It will still generate Trace IDs for `slog`, but send zero network traffic.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	log.Println("OpenTelemetry initialized in Headless Mode (Trace IDs only, no Jaeger)")
	return tp
}
