package model

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/rs/zerolog"
	"net/http"
	"noty/pkg/logging"
	"time"
)

// Sending keeps sending data.
type (
	Filter struct {
		Tags  []string `json:"tags,omitempty"`
		Codes []int    `json:"codes,omitempty"`
	}
	Sending struct {
		ID      uuid.UUID `json:"id"`
		StartAt time.Time `json:"start_at,omitempty"`
		Text    string    `json:"text"`
		Filter  Filter    `json:"filter,omitempty"`
		StopAt  time.Time `json:"stop_at,omitempty"`
	}
	Sendings []*Sending

	SendingStatus struct {
		Sending *Sending `json:"sending"`
		New     int      `json:"new"`
		Sent    int      `json:"sent"`
	}
	SendingsStatus []*SendingStatus
)

func (dst *Filter) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return errors.New("NULL values can't be decoded. Scan into a &*MyType to handle NULLs")
	}

	if err := (pgtype.CompositeFields{&dst.Tags, &dst.Codes}).DecodeBinary(ci, src); err != nil {
		return err
	}

	return nil
}

func (src Filter) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	var a pgtype.TextArray
	err = a.Set(src.Tags)
	if err != nil {
		return nil, err
	}

	var b pgtype.Int8Array
	err = b.Set(src.Codes)
	if err != nil {
		return nil, err
	}

	return (pgtype.CompositeFields{&a, &b}).EncodeBinary(ci, buf)
}

func (*Sending) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (Sendings) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*SendingsStatus) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Sending) Bind(r *http.Request) error {
	//if s.ID == uuid.Nil {
	//	s.ID, _ = uuid.NewUUID()
	//}

	if s.Text == "" {
		return fmt.Errorf("text is a required field")
	}

	if s.StartAt.IsZero() {
		return fmt.Errorf("start_at is a required field")
	}
	if s.StopAt.IsZero() {
		return fmt.Errorf("stop_at is a required field")
	}
	return nil
}

// GetLoggerContext enriches logger context with essential fields.
func (s *Sending) GetLoggerContext(logCtx zerolog.Context) zerolog.Context {
	if s.ID != uuid.Nil {
		logCtx = logCtx.Str(logging.SendingIDKey, s.ID.String())
	}

	return logCtx
}
