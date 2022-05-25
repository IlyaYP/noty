package model

import (
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Message struct {
	ID        int64         `json:"id" yaml:"id"`
	CreatedAt time.Time     `json:"created_at" yaml:"created_at"`
	Status    MessageStatus `json:"status" yaml:"status"`
	SendingID uuid.UUID     `json:"sending_id" yaml:"sending_id"`
	ClientID  uuid.UUID     `json:"client_id" yaml:"client_id"`
}

type MessageStatus string

const (
	MessageStatusCreated   MessageStatus = "CREATED"
	MessageStatusShipped   MessageStatus = "SHIPPED"
	MessageStatusDelivered MessageStatus = "DELIVERED"
	MessageStatusCompleted MessageStatus = "COMPLETED"
)

func (*Message) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
