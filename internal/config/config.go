package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel        string        `env:"LOG_LEVEL" env-required:"true"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-required:"true"`
	IdempotencyTTL  time.Duration `env:"IDEMPOTENCY_TTL" env-default:"24h"`
	Http            HttpConfig    `env-prefix:"HTTP_"`
	DB              DBConfig      `env-prefix:"DATABASE_"`
	OTel            OTelConfig    `env-prefix:"OTEL_"`
}

type HttpConfig struct {
	Addr              string        `env:"ADDR" env-required:"true"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT" env-default:"30s"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" env-default:"5s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT" env-default:"30s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

type DBConfig struct {
	Url             string        `env:"URL" env-required:"true"`
	MaxConns        int32         `env:"MAX_CONNS" env-required:"true"`
	MinConns        int32         `env:"MIN_CONNS" env-required:"true"`
	ConnectTimeout  time.Duration `env:"CONNECT_TIMEOUT" env-required:"true"`
	QueryTimeout    time.Duration `env:"QUERY_TIMEOUT" env-required:"true"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME" env-required:"true"`
}

type OTelConfig struct {
	ServiceName        string `env:"SERVICE_NAME" env-required:"true"`
	ExporterEndpoint   string `env:"EXPORTER_OTLP_ENDPOINT" env-required:"true"`
	ExporterProtocol   string `env:"EXPORTER_OTLP_PROTOCOL" env-required:"true"`
	ResourceAttributes string `env:"RESOURCE_ATTRIBUTES" env-required:"true"`
	TracesSampler      string `env:"TRACES_SAMPLER" env-required:"true"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}
