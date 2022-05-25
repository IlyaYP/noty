package model

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"net/http"
	"noty/pkg/logging"
)

// Client keeps client data.
type Client struct {
	ID           uuid.UUID `json:"id" yaml:"id"`
	PhoneNumber  int       `json:"phone_number" yaml:"phone_number"`
	MobileOpCode string    `json:"mobile_op_code" yaml:"mobile_op_code"`
	Tag          string    `json:"tag" yaml:"tag"`
	TZ           string    `json:"tz" yaml:"tz"`
}

func (*Client) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (c *Client) Bind(r *http.Request) error {
	if c.PhoneNumber == 0 {
		return fmt.Errorf("PhoneNumber is a required field")
	}
	return nil
}

// GetLoggerContext enriches logger context with essential Client fields.
func (c *Client) GetLoggerContext(logCtx zerolog.Context) zerolog.Context {
	logCtx = logCtx.Int("phone", c.PhoneNumber)

	if c.ID != uuid.Nil {
		logCtx = logCtx.Str(logging.ClientIDKey, c.ID.String())
	}

	return logCtx
}
