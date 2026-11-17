package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func (c *Config) validate() error {
	return errors.Join(
		validateLogLevel(c.LogLevel),
		validatePositiveDuration(envShutdownTimeout, c.ShutdownTimeout),
		validatePositiveDuration(envIdempotencyTTL, c.IdempotencyTTL),
		c.Http.validate(),
		c.DB.validate(),
	)
}

func (c HttpConfig) validate() error {
	return errors.Join(
		validateNonEmptyString(envHTTPAddr, c.Addr),
		validatePositiveDuration(envHTTPReadTimeout, c.ReadTimeout),
		validatePositiveDuration(envHTTPReadHeaderTimeout, c.ReadHeaderTimeout),
		validatePositiveDuration(envHTTPWriteTimeout, c.WriteTimeout),
		validatePositiveDuration(envHTTPIdleTimeout, c.IdleTimeout),
	)
}

func (c DBConfig) validate() error {
	var maxConnsErr error
	if c.MaxConns <= 0 {
		maxConnsErr = fmt.Errorf("%s must be positive", envDatabaseMaxConns)
	}

	var minConnsErr error
	if c.MinConns <= 0 {
		minConnsErr = fmt.Errorf("%s must be positive", envDatabaseMinConns)
	}

	var connsRangeErr error
	if c.MaxConns < c.MinConns {
		connsRangeErr = fmt.Errorf("%s must not exceed %s", envDatabaseMinConns, envDatabaseMaxConns)
	}

	return errors.Join(
		validateNonEmptyString(envDatabaseURL, c.Url),
		maxConnsErr,
		minConnsErr,
		connsRangeErr,
		validatePositiveDuration(envDatabaseConnectTimeout, c.ConnectTimeout),
		validatePositiveDuration(envDatabaseQueryTimeout, c.QueryTimeout),
		validatePositiveDuration(envDatabaseMaxConnLifetime, c.MaxConnLifetime),
	)
}

func validateLogLevel(value string) error {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("%s must be one of: debug, info, warn, error", envLogLevel)
	}
}

func validateNonEmptyString(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", name)
	}

	return nil
}

func validatePositiveDuration(name string, value time.Duration) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", name)
	}

	return nil
}
