package model

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"net/http"
	"noty/pkg/logging"
	"time"
)

// Sending keeps sending data.
type Sending struct {
	ID      uuid.UUID `json:"id" yaml:"id"`
	StartAt time.Time `json:"start_at" yaml:"start_at"`
	Text    string    `json:"text" yaml:"text"`
	Filter  string    `json:"filter" yaml:"filter"` // TODO: invent something instead of string
	StopAt  time.Time `json:"stop_at" yaml:"stop_at"`
}

func (*Sending) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Sending) Bind(r *http.Request) error {
	if s.Text == "" {
		return fmt.Errorf("text is a required field")
	}
	return nil
}

// GetLoggerContext enriches logger context with essential Client fields.
func (s *Sending) GetLoggerContext(logCtx zerolog.Context) zerolog.Context {
	if s.ID != uuid.Nil {
		logCtx = logCtx.Str(logging.SendingIDKey, s.ID.String())
	}

	return logCtx
}
