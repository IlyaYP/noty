package model

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"net/http"
	"noty/pkg/logging"
	"time"
)

type (
	Message struct {
		ID        int64         `json:"id" yaml:"id"`
		CreatedAt time.Time     `json:"created_at" yaml:"created_at"`
		Status    MessageStatus `json:"status" yaml:"status"`
		SendingID uuid.UUID     `json:"sending_id" yaml:"sending_id"`
		ClientID  uuid.UUID     `json:"client_id" yaml:"client_id"`
	}

	Messages []*Message

	MessageToSend struct {
		ID    int64  `json:"id" yaml:"id"`
		Phone int    `json:"phone" yaml:"phone"`
		Text  string `json:"text" yaml:"text"`
	}

	MessageStatus string
)

func (m MessageToSend) Bind(r *http.Request) error {
	//TODO implement me
	panic("implement me")
}

func (*Message) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*Messages) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// GetLoggerContext enriches logger context with essential fields.
func (m *Message) GetLoggerContext(logCtx zerolog.Context) zerolog.Context {
	if m.ID != 0 {
		logCtx = logCtx.Int64(logging.MessageIDKey, m.ID)
	}

	return logCtx
}

const (
	MessageStatusNew  MessageStatus = "NEW"
	MessageStatusSent MessageStatus = "SENT"
)

var (
	// messageStatusMap maps OrderStatus value to its int representation.
	messageStatusToIntMap = map[MessageStatus]int{
		MessageStatusNew:  1,
		MessageStatusSent: 2,
	}

	// messageStatusToStrMap maps OrderStatus value to its string representation.
	messageStatusToStrMap = map[int]MessageStatus{
		1: MessageStatusNew,
		2: MessageStatusSent,
	}
)

// NewMessageStatusFromInt returns OrderStatus by its int representation (might be invalid).
func NewMessageStatusFromInt(v int) MessageStatus {
	return messageStatusToStrMap[v]
}

// String implements the fmt.Stringer interface.
func (s MessageStatus) String() string {
	return string(s)
}

// Int returns enum value int representation.
func (s MessageStatus) Int() int {
	return messageStatusToIntMap[s]
}

// Validate validates enum value.
func (s MessageStatus) Validate() error {
	_, found := messageStatusToIntMap[s]
	if !found {
		return fmt.Errorf("unknown value: %v", s)
	}

	return nil
}
