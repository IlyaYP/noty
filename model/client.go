package model

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"net/http"
	"noty/pkg/logging"
)

// Client keeps client data.
type (
	Client struct {
		ID     uuid.UUID `json:"id" yaml:"id"`
		Phone  int       `json:"phone" yaml:"number"`
		OpCode int       `json:"op_code" yaml:"op_code"`
		Tag    string    `json:"tag" yaml:"tag"`
		TZ     string    `json:"tz" yaml:"tz"`
	}
	Clients []Client
)

func (Clients) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*Client) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (c *Client) Bind(r *http.Request) error {
	if c.Phone == 0 {
		return fmt.Errorf("phone is a required field")
	}

	//if c.OpCode == "" {
	//	return fmt.Errorf("op_code is a required field")
	//}

	//if c.ID == uuid.Nil {
	//	c.ID, _ = uuid.NewUUID()
	//}

	return nil
}

// GetLoggerContext enriches logger context with essential Client fields.
func (c *Client) GetLoggerContext(logCtx zerolog.Context) zerolog.Context {
	logCtx = logCtx.Int("phone", c.Phone)

	if c.ID != uuid.Nil {
		logCtx = logCtx.Str(logging.ClientIDKey, c.ID.String())
	}

	return logCtx
}
