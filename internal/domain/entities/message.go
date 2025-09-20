package entities

import (
	"time"

	"github.com/google/uuid"
)

type MessageStatus string

const (
	MessageStatusPending MessageStatus = "pending"
	MessageStatusSent    MessageStatus = "sent"
	MessageStatusFailed  MessageStatus = "failed"
)

type Message struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	Content     string        `json:"content" db:"content"`
	PhoneNumber string        `json:"phone_number" db:"phone_number"`
	Status      MessageStatus `json:"status" db:"status"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
	SentAt      *time.Time    `json:"sent_at,omitempty" db:"sent_at"`

	ExternalMessageID *string `json:"external_message_id,omitempty" db:"external_message_id"`
	ErrorMessage      *string `json:"error_message,omitempty" db:"error_message"`
}

func (m *Message) Validate() error {
	if m.Content == "" {
		return ErrInvalidMessageContent
	}

	if len(m.Content) > 160 { // SMS character limit
		return ErrMessageTooLong
	}

	if m.PhoneNumber == "" {
		return ErrInvalidPhoneNumber
	}

	return nil
}

func (m *Message) MarkAsSent(externalMessageID string) {
	now := time.Now()
	m.Status = MessageStatusSent
	m.SentAt = &now
	m.UpdatedAt = now
	m.ExternalMessageID = &externalMessageID
}

func (m *Message) MarkAsFailed(errorMsg string) {
	m.Status = MessageStatusFailed
	m.UpdatedAt = time.Now()
	m.ErrorMessage = &errorMsg
}

func (m *Message) IsPending() bool {
	return m.Status == MessageStatusPending
}

func (m *Message) IsSent() bool {
	return m.Status == MessageStatusSent
}
