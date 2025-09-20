package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"message-sending-service/internal/application/dto"
	"message-sending-service/internal/domain/entities"
	domainUsecases "message-sending-service/internal/domain/usecases"
)

// mock testler
type mockMessageUseCase struct {
	createMessageFunc   func(ctx context.Context, content, phoneNumber string) (*entities.Message, error)
	getMessageByIDFunc  func(ctx context.Context, id uuid.UUID) (*entities.Message, error)
	getSentMessagesFunc func(ctx context.Context, page, limit int) ([]*entities.Message, int64, error)
	getMessageStatsFunc func(ctx context.Context) (*domainUsecases.MessageStats, error)
	sendMessageFunc     func(ctx context.Context, message *entities.Message) error
}

func (m *mockMessageUseCase) CreateMessage(ctx context.Context, content, phoneNumber string) (*entities.Message, error) {
	if m.createMessageFunc != nil {
		return m.createMessageFunc(ctx, content, phoneNumber)
	}
	return &entities.Message{
		ID:          uuid.New(),
		Content:     content,
		PhoneNumber: phoneNumber,
		Status:      entities.MessageStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *mockMessageUseCase) GetMessageByID(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
	if m.getMessageByIDFunc != nil {
		return m.getMessageByIDFunc(ctx, id)
	}
	return &entities.Message{
		ID:          id,
		Content:     "Test message",
		PhoneNumber: "+1234567890",
		Status:      entities.MessageStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *mockMessageUseCase) GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error) {
	return nil, nil
}

func (m *mockMessageUseCase) GetSentMessages(ctx context.Context, page, limit int) ([]*entities.Message, int64, error) {
	if m.getSentMessagesFunc != nil {
		return m.getSentMessagesFunc(ctx, page, limit)
	}
	return []*entities.Message{}, 0, nil
}

func (m *mockMessageUseCase) SendMessage(ctx context.Context, message *entities.Message) error {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, message)
	}
	return nil
}

func (m *mockMessageUseCase) ProcessPendingMessages(ctx context.Context, batchSize int) (int, error) {
	return 0, nil
}

func (m *mockMessageUseCase) GetMessageStats(ctx context.Context) (*domainUsecases.MessageStats, error) {
	if m.getMessageStatsFunc != nil {
		return m.getMessageStatsFunc(ctx)
	}
	return &domainUsecases.MessageStats{
		TotalMessages:   10,
		PendingMessages: 5,
		SentMessages:    3,
		FailedMessages:  2,
	}, nil
}

func TestMessageHandler_CreateMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, content, phoneNumber string) (*entities.Message, error)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful creation",
			requestBody: dto.CreateMessageRequest{
				Content:     "Test message",
				PhoneNumber: "+1234567890",
			},
			mockFunc:       nil, // Use default mock
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "empty content",
			requestBody: dto.CreateMessageRequest{
				Content:     "",
				PhoneNumber: "+1234567890",
			},
			mockFunc: func(ctx context.Context, content, phoneNumber string) (*entities.Message, error) {
				return nil, entities.ErrInvalidMessageContent
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid json",
			requestBody:    "invalid json",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "message too long",
			requestBody: dto.CreateMessageRequest{
				Content:     string(make([]byte, 161)),
				PhoneNumber: "+1234567890",
			},
			mockFunc: func(ctx context.Context, content, phoneNumber string) (*entities.Message, error) {
				return nil, entities.ErrMessageTooLong
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mockMessageUseCase{
				createMessageFunc: tt.mockFunc,
			}
			logger, _ := zap.NewNop(), zap.NewNop()
			handler := NewMessageHandler(mockUseCase, logger)

			var reqBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/messages", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.CreateMessage(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectError {
				if _, exists := response["error"]; !exists {
					t.Error("Expected error in response but got none")
				}
			} else {
				if _, exists := response["data"]; !exists {
					t.Error("Expected data in response but got none")
				}
			}
		})
	}
}

func TestMessageHandler_GetMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		messageID      string
		mockFunc       func(ctx context.Context, id uuid.UUID) (*entities.Message, error)
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful get",
			messageID:      uuid.New().String(),
			mockFunc:       nil,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "invalid uuid",
			messageID: "invalid-uuid",
			mockFunc: func(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
				return nil, entities.ErrMessageNotFound
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "message not found",
			messageID: uuid.New().String(),
			mockFunc: func(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
				return nil, entities.ErrMessageNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mockMessageUseCase{
				getMessageByIDFunc: tt.mockFunc,
			}
			logger, _ := zap.NewNop(), zap.NewNop()
			handler := NewMessageHandler(mockUseCase, logger)

			req := httptest.NewRequest("GET", "/messages/"+tt.messageID, nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "id", Value: tt.messageID},
			}

			handler.GetMessage(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectError {
				if _, exists := response["error"]; !exists {
					t.Error("Expected error in response but got none")
				}
			} else {
				if _, exists := response["data"]; !exists {
					t.Error("Expected data in response but got none")
				}
			}
		})
	}
}

func TestMessageHandler_GetSentMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUseCase := &mockMessageUseCase{
		getSentMessagesFunc: func(ctx context.Context, page, limit int) ([]*entities.Message, int64, error) {
			messages := []*entities.Message{
				{
					ID:          uuid.New(),
					Content:     "Test message 1",
					PhoneNumber: "+1234567890",
					Status:      entities.MessageStatusSent,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			}
			return messages, 1, nil
		},
	}

	logger, _ := zap.NewNop(), zap.NewNop()
	handler := NewMessageHandler(mockUseCase, logger)

	req := httptest.NewRequest("GET", "/messages/sent?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetSentMessages(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if _, exists := response["data"]; !exists {
		t.Error("Expected data in response but got none")
	}
}

func TestMessageHandler_GetMessageStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUseCase := &mockMessageUseCase{} // Use default mock
	logger, _ := zap.NewNop(), zap.NewNop()
	handler := NewMessageHandler(mockUseCase, logger)

	req := httptest.NewRequest("GET", "/messages/stats", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetMessageStats(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, exists := response["data"]
	if !exists {
		t.Error("Expected data in response but got none")
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Error("Expected data to be an object")
	}

	expectedFields := []string{"total_messages", "pending_messages", "sent_messages", "failed_messages"}
	for _, field := range expectedFields {
		if _, exists := dataMap[field]; !exists {
			t.Errorf("Expected field %s in stats data", field)
		}
	}
}
