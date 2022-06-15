package sender

import (
	"fmt"
	"time"
)

const (
	defaultAddress       = "https://probe.fbrq.cloud/v1" //"localhost:8081"
	defaultConfigTimeOut = 10
	defaultToken         = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODQ4NTgyOTIsImlzcyI6ImZhYnJpcXVlIiwibmFtZSI6IkBzaXNfdCJ9.ZBRVHs0Rqgod4ayfVTnt75fZi3NPr70fCYWHtG76iiY"
	defaultEndpoint      = "/send"
)

type Config struct {
	Address  string
	timeout  time.Duration
	Token    string
	endPoint string
}

// validate performs a basic validation.
func (c Config) validate() error {
	if c.Address == "" {
		return fmt.Errorf("%s field: empty", "ACCRUAL_SYSTEM_ADDRESS")
	}
	if c.timeout == 0 {
		return fmt.Errorf("%s field: empty", "timeout")
	}

	return nil
}

// NewDefaultConfig builds a Config with default values.
func NewDefaultConfig() Config {
	return Config{
		Address:  defaultAddress,
		timeout:  time.Duration(defaultConfigTimeOut) * time.Second,
		Token:    defaultToken,
		endPoint: defaultAddress + defaultEndpoint,
	}
}
