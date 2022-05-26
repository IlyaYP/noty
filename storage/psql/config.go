package psql

import (
	"fmt"
	"time"
)

const (
	defaultConfigEndpoint = "postgres://postgres:postgres@localhost:5432/noty"
	defaultConfigTimeOut  = 1
)

type Config struct {
	DSN     string `env:"DATABASE_URI"`
	timeout time.Duration
}

// validate performs a basic validation.
func (c Config) validate() error {
	if c.DSN == "" {
		return fmt.Errorf("%s field: empty", "DATABASE_URI")
	}
	if c.timeout == 0 {
		return fmt.Errorf("%s field: empty", "timeout")
	}

	return nil
}

// NewDefaultConfig builds a Config with default values.
func NewDefaultConfig() Config {
	return Config{
		DSN:     defaultConfigEndpoint,
		timeout: time.Duration(defaultConfigTimeOut) * time.Second,
	}
}
