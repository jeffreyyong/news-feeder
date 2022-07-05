package apppostgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	ddsql "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
)

const (
	driverName                  = "postgres"
	defaultDatadogAnalyticsRate = 0.25
)

// Config to connect to postgres database.
type Config struct {
	tracingProvider      trace.TracerProvider
	datadogAnalyticsRate float64
}

// Option is an option
type Option func(o *Config)

// NewBasicClient creates a new postgres driver, without an app dependency.
// Caution: This is intended only for CLI tools and libraries which have a non-app service use case.
// In the majority of cases the `New()` function should be used so you get best practices for free.
func NewBasicClient(ctx context.Context, serviceName, dsn string, opts ...Option) (*sql.DB, error) {
	var c Config
	for _, opt := range opts {
		opt(&c)
	}

	otelOptions := []otelsql.Option{
		otelsql.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	}
	if c.tracingProvider != nil {
		otelOptions = append(otelOptions, otelsql.WithTracerProvider(c.tracingProvider))
	}

	datadogAnalyticsRate := defaultDatadogAnalyticsRate
	if c.datadogAnalyticsRate > 0 {
		datadogAnalyticsRate = c.datadogAnalyticsRate
	}

	driver := pq.Driver{}
	otelDriver := otelsql.WrapDriver(&driver, otelOptions...)
	ddsql.Register(driverName, otelDriver,
		ddsql.WithServiceName(serviceName),
		ddsql.WithAnalyticsRate(datadogAnalyticsRate),
	)

	client, err := ddsql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres_unable_to_connect: %w", err)
	}

	if err := client.PingContext(ctx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("postgres_unable_to_ping: %w", err)
	}
	return client, nil
}
