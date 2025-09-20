package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"message-sending-service/internal/domain/entities"
	"message-sending-service/internal/infrastructure/config"
	"message-sending-service/internal/infrastructure/external"
)

func TestMessageUseCase_CreateMessage_Simple(t *testing.T) {
	cfg := &config.Config{
		External: config.ExternalConfig{
			MessageAPIURL: "https://webhook.site/test",
			Timeout:       5 * time.Second,
		},
	}

	apiClient := external.NewMessageAPIClient(cfg)
	logger, _ := zap.NewNop(), zap.NewNop()

	useCase := NewMessageUseCase(nil, nil, apiClient, logger)

	tests := []struct {
		name        string
		content     string
		phoneNumber string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "empty content",
			content:     "",
			phoneNumber: "+1234567890",
			wantErr:     true,
			expectedErr: entities.ErrInvalidMessageContent,
		},
		{
			name:        "empty phone number",
			content:     "Test message",
			phoneNumber: "",
			wantErr:     true,
			expectedErr: entities.ErrInvalidPhoneNumber,
		},
		{
			name:        "message too long",
			content:     string(make([]byte, 161)),
			phoneNumber: "+1234567890",
			wantErr:     true,
			expectedErr: entities.ErrMessageTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			message := &entities.Message{
				ID:          uuid.New(),
				Content:     tt.content,
				PhoneNumber: tt.phoneNumber,
				Status:      entities.MessageStatusPending,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			err := message.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			_, err = useCase.CreateMessage(ctx, tt.content, tt.phoneNumber)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			}
		})
	}
}

func TestMessage_BusinessLogic(t *testing.T) {
	t.Run("valid message creation", func(t *testing.T) {
		message := &entities.Message{
			ID:          uuid.New(),
			Content:     "Valid test message",
			PhoneNumber: "+1234567890",
			Status:      entities.MessageStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := message.Validate()
		if err != nil {
			t.Errorf("Expected no error for valid message, got: %v", err)
		}

		if !message.IsPending() {
			t.Error("Expected message to be pending")
		}
	})

	t.Run("mark message as sent", func(t *testing.T) {
		message := &entities.Message{
			ID:          uuid.New(),
			Content:     "Test message",
			PhoneNumber: "+1234567890",
			Status:      entities.MessageStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		externalID := "ext_123"
		message.MarkAsSent(externalID)

		if !message.IsSent() {
			t.Error("Expected message to be marked as sent")
		}

		if message.ExternalMessageID == nil || *message.ExternalMessageID != externalID {
			t.Errorf("Expected external ID to be %s", externalID)
		}

		if message.SentAt == nil {
			t.Error("Expected SentAt to be set")
		}
	})

	t.Run("mark message as failed", func(t *testing.T) {
		message := &entities.Message{
			ID:          uuid.New(),
			Content:     "Test message",
			PhoneNumber: "+1234567890",
			Status:      entities.MessageStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		errorMsg := "API Error"
		message.MarkAsFailed(errorMsg)

		if message.Status != entities.MessageStatusFailed {
			t.Error("Expected message to be marked as failed")
		}

		if message.ErrorMessage == nil || *message.ErrorMessage != errorMsg {
			t.Errorf("Expected error message to be %s", errorMsg)
		}
	})
}

func TestScheduler_BusinessLogic(t *testing.T) {
	t.Run("scheduler info management", func(t *testing.T) {
		info := &entities.SchedulerInfo{
			Status:        entities.SchedulerStatusStopped,
			MessagesCount: 0,
			Interval:      2 * time.Minute,
			BatchSize:     2,
		}

		if info.IsRunning() {
			t.Error("Expected scheduler to not be running initially")
		}

		info.Start()
		if !info.IsRunning() {
			t.Error("Expected scheduler to be running after start")
		}

		if info.NextRun == nil {
			t.Error("Expected NextRun to be set after start")
		}

		info.Stop()
		if info.IsRunning() {
			t.Error("Expected scheduler to not be running after stop")
		}

		if info.NextRun != nil {
			t.Error("Expected NextRun to be nil after stop")
		}
	})
}

func BenchmarkMessage_Validate(b *testing.B) {
	message := &entities.Message{
		Content:     "This is a benchmark test message for validation performance",
		PhoneNumber: "+1234567890",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message.Validate()
	}
}

func BenchmarkMessage_MarkAsSent(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message := &entities.Message{
			ID:        uuid.New(),
			Status:    entities.MessageStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		message.MarkAsSent("ext_123")
	}
}
