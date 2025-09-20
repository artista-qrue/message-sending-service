package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		wantErr error
	}{
		{
			name: "valid message",
			message: Message{
				Content:     "Hello, this is a test message",
				PhoneNumber: "+1234567890",
			},
			wantErr: nil,
		},
		{
			name: "empty content",
			message: Message{
				Content:     "",
				PhoneNumber: "+1234567890",
			},
			wantErr: ErrInvalidMessageContent,
		},
		{
			name: "empty phone number",
			message: Message{
				Content:     "Hello",
				PhoneNumber: "",
			},
			wantErr: ErrInvalidPhoneNumber,
		},
		{
			name: "message too long",
			message: Message{
				Content:     string(make([]byte, 161)), // 161 characters
				PhoneNumber: "+1234567890",
			},
			wantErr: ErrMessageTooLong,
		},
		{
			name: "exactly 160 characters",
			message: Message{
				Content:     string(make([]byte, 160)), // 160 characters
				PhoneNumber: "+1234567890",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if err != tt.wantErr {
				t.Errorf("Message.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessage_MarkAsSent(t *testing.T) {
	message := &Message{
		ID:        uuid.New(),
		Status:    MessageStatusPending,
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	externalID := "ext_123"
	oldUpdatedAt := message.UpdatedAt

	message.MarkAsSent(externalID)

	if message.Status != MessageStatusSent {
		t.Errorf("Expected status to be %v, got %v", MessageStatusSent, message.Status)
	}

	if message.ExternalMessageID == nil || *message.ExternalMessageID != externalID {
		t.Errorf("Expected external message ID to be %v, got %v", externalID, message.ExternalMessageID)
	}

	if message.SentAt == nil {
		t.Error("Expected SentAt to be set")
	}

	if !message.UpdatedAt.After(oldUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestMessage_MarkAsFailed(t *testing.T) {
	message := &Message{
		ID:        uuid.New(),
		Status:    MessageStatusPending,
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	errorMsg := "Network error"
	oldUpdatedAt := message.UpdatedAt

	message.MarkAsFailed(errorMsg)

	if message.Status != MessageStatusFailed {
		t.Errorf("Expected status to be %v, got %v", MessageStatusFailed, message.Status)
	}

	if message.ErrorMessage == nil || *message.ErrorMessage != errorMsg {
		t.Errorf("Expected error message to be %v, got %v", errorMsg, message.ErrorMessage)
	}

	if !message.UpdatedAt.After(oldUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestMessage_IsPending(t *testing.T) {
	message := &Message{Status: MessageStatusPending}
	if !message.IsPending() {
		t.Error("Expected message to be pending")
	}

	message.Status = MessageStatusSent
	if message.IsPending() {
		t.Error("Expected message to not be pending")
	}
}

func TestMessage_IsSent(t *testing.T) {
	message := &Message{Status: MessageStatusSent}
	if !message.IsSent() {
		t.Error("Expected message to be sent")
	}

	message.Status = MessageStatusPending
	if message.IsSent() {
		t.Error("Expected message to not be sent")
	}
}
