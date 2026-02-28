package observability

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	ServiceName    string
	OTLPEndpoint   string
	OTLPInsecure   bool
	TraceEnabled   bool
	MetricsEnabled bool
}

func LoadConfig(defaultServiceName string) Config {
	serviceName := envOr("OTEL_SERVICE_NAME", defaultServiceName)

	endpoint := firstNonEmpty(
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"),
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	)

	insecure := envBool("OTEL_EXPORTER_OTLP_INSECURE", true)
	traceEnabled := envBool("OTEL_TRACES_ENABLED", true)
	metricsEnabled := envBool("OTEL_METRICS_ENABLED", true)

	if endpoint == "" {
		traceEnabled = false
	}

	if parsedHost, secure, ok := normalizeEndpoint(endpoint); ok {
		endpoint = parsedHost
		if secure {
			insecure = false
		}
	}

	return Config{
		ServiceName:    serviceName,
		OTLPEndpoint:   endpoint,
		OTLPInsecure:   insecure,
		TraceEnabled:   traceEnabled,
		MetricsEnabled: metricsEnabled,
	}
}

func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			attribute.String("service.name", cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	shutdowns := make([]func(context.Context) error, 0, 2)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	if cfg.MetricsEnabled {
		metricExporter, err := prometheus.New()
		if err != nil {
			return nil, err
		}
		meterProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(metricExporter),
		)
		otel.SetMeterProvider(meterProvider)
		shutdowns = append(shutdowns, meterProvider.Shutdown)
	}

	if cfg.TraceEnabled {
		options := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		}
		if cfg.OTLPInsecure {
			options = append(options, otlptracegrpc.WithInsecure())
		}

		traceExporter, err := otlptracegrpc.New(ctx, options...)
		if err != nil {
			return nil, err
		}

		traceProvider := sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(traceExporter),
		)
		otel.SetTracerProvider(traceProvider)
		shutdowns = append(shutdowns, traceProvider.Shutdown)
	}

	return func(ctx context.Context) error {
		var combined error
		for i := len(shutdowns) - 1; i >= 0; i-- {
			combined = errors.Join(combined, shutdowns[i](ctx))
		}
		return combined
	}, nil
}

func WrapHTTPHandler(h http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(h, operation)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func LogStartup(cfg Config) {
	log.Printf(
		"observability initialized service=%s traces_enabled=%t metrics_enabled=%t otlp_endpoint=%q metrics_path=/metrics",
		cfg.ServiceName,
		cfg.TraceEnabled,
		cfg.MetricsEnabled,
		cfg.OTLPEndpoint,
	)
}

func ShutdownWithTimeout(shutdown func(context.Context) error, timeout time.Duration) {
	if shutdown == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := shutdown(ctx); err != nil {
		log.Printf("observability shutdown error: %v", err)
	}
}

func envOr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func normalizeEndpoint(v string) (host string, secure bool, ok bool) {
	u, err := url.Parse(v)
	if err != nil || u.Host == "" || u.Scheme == "" {
		return "", false, false
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return u.Host, true, true
	case "http":
		return u.Host, false, true
	default:
		return "", false, false
	}
}
