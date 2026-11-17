package config

const (
	envLogLevel                = "LOG_LEVEL"
	envShutdownTimeout         = "SHUTDOWN_TIMEOUT"
	envIdempotencyTTL          = "IDEMPOTENCY_TTL"
	envHTTPAddr                = "HTTP_ADDR"
	envHTTPReadTimeout         = "HTTP_READ_TIMEOUT"
	envHTTPReadHeaderTimeout   = "HTTP_READ_HEADER_TIMEOUT"
	envHTTPWriteTimeout        = "HTTP_WRITE_TIMEOUT" // #nosec G101 -- environment variable name, not a credential.
	envHTTPIdleTimeout         = "HTTP_IDLE_TIMEOUT"
	envDatabaseURL             = "DATABASE_URL"
	envDatabaseMinConns        = "DATABASE_MIN_CONNS"
	envDatabaseMaxConns        = "DATABASE_MAX_CONNS"
	envDatabaseConnectTimeout  = "DATABASE_CONNECT_TIMEOUT"
	envDatabaseQueryTimeout    = "DATABASE_QUERY_TIMEOUT"
	envDatabaseMaxConnLifetime = "DATABASE_MAX_CONN_LIFETIME"
	envOtelServiceName         = "OTEL_SERVICE_NAME"
	envOtelExporterEndpoint    = "OTEL_EXPORTER_OTLP_ENDPOINT"
	envOtelExporterProtocol    = "OTEL_EXPORTER_OTLP_PROTOCOL"
	envOtelResourceAttributes  = "OTEL_RESOURCE_ATTRIBUTES"
	envOtelTracesSampler       = "OTEL_TRACES_SAMPLER"
)
