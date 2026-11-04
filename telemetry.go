package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// initTelemetry inicializa o SDK do OpenTelemetry (traces e metricas).
// Retorna uma funcao de shutdown que deve ser deferida no main().
// Se o OTel Collector estiver inacessivel, loga o aviso e retorna um no-op,
// garantindo que o servico suba normalmente mesmo sem telemetria.
//
// IMPORTANTE: deve ser chamada ANTES de otelhttp.NewHandler, pois o handler
// captura o MeterProvider global no momento da criacao.
func initTelemetry(ctx context.Context) func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector.observability.svc.cluster.local:4317"
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "auth-service"
	}

	environment := os.Getenv("DEPLOYMENT_ENVIRONMENT")
	if environment == "" {
		environment = "dev"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		log.Printf("[OTel] Erro ao criar resource: %v. Telemetria desabilitada.", err)
		return func(context.Context) error { return nil }
	}

	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("[OTel] Aviso: OTel Collector inacessivel em %s: %v. Telemetria desabilitada.", endpoint, err)
		return func(context.Context) error { return nil }
	}

	var shutdownFns []func(context.Context) error

	// --- Traces ---
	traceExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Printf("[OTel] Erro ao criar trace exporter: %v", err)
	} else {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExp),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		otel.SetTracerProvider(tp)
		shutdownFns = append(shutdownFns, tp.Shutdown)
	}

	// Propagador W3C Trace Context — essencial para trace distribuido
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// --- Metricas ---
	// O otelhttp registra automaticamente http.server.duration (ms) com o
	// atributo http.status_code; o Collector encaminha ao Prometheus.
	metricExp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		log.Printf("[OTel] Erro ao criar metric exporter: %v", err)
	} else {
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
				sdkmetric.WithInterval(15*time.Second),
			)),
			sdkmetric.WithResource(res),
		)
		otel.SetMeterProvider(mp)
		shutdownFns = append(shutdownFns, mp.Shutdown)
	}

	log.Printf("[OTel] Telemetria (traces + metricas) inicializada — servico: %s, endpoint: %s", serviceName, endpoint)

	return func(ctx context.Context) error {
		for _, fn := range shutdownFns {
			if err := fn(ctx); err != nil {
				log.Printf("[OTel] Erro no shutdown: %v", err)
			}
		}
		return nil
	}
}
