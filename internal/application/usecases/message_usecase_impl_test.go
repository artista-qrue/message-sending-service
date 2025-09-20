package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"message-sending-service/internal/domain/entities"
	"message-sending-service/internal/infrastructure/external"
)

type mockMessageRepository struct {
	messages     map[uuid.UUID]*entities.Message
	pendingCount int
	shouldFail   bool
}

func newMockMessageRepository() *mockMessageRepository {
	return &mockMessageRepository{
		messages: make(map[uuid.UUID]*entities.Message),
	}
}

func (m *mockMessageRepository) Create(ctx context.Context, message *entities.Message) error {
	if m.shouldFail {
		return errors.New("database error")
	}
	m.messages[message.ID] = message
	return nil
}

func (m *mockMessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
	if m.shouldFail {
		return nil, errors.New("database error")
	}
	if msg, exists := m.messages[id]; exists {
		return msg, nil
	}
	return nil, entities.ErrMessageNotFound
}

func (m *mockMessageRepository) Update(ctx context.Context, message *entities.Message) error {
	if m.shouldFail {
		return errors.New("database error")
	}
	m.messages[message.ID] = message
	return nil
}

func (m *mockMessageRepository) GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error) {
	if m.shouldFail {
		return nil, errors.New("database error")
	}

	var pending []*entities.Message
	for _, msg := range m.messages {
		if msg.Status == entities.MessageStatusPending {
			pending = append(pending, msg)
			if len(pending) >= limit {
				break
			}
		}
	}
	return pending, nil
}

func (m *mockMessageRepository) GetSentMessages(ctx context.Context, offset, limit int) ([]*entities.Message, error) {
	if m.shouldFail {
		return nil, errors.New("database error")
	}

	var sent []*entities.Message
	for _, msg := range m.messages {
		if msg.Status == entities.MessageStatusSent {
			sent = append(sent, msg)
		}
	}
	return sent, nil
}

func (m *mockMessageRepository) CountByStatus(ctx context.Context, status entities.MessageStatus) (int64, error) {
	if m.shouldFail {
		return 0, errors.New("database error")
	}

	count := int64(0)
	for _, msg := range m.messages {
		if msg.Status == status {
			count++
		}
	}
	return count, nil
}

func (m *mockMessageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.shouldFail {
		return errors.New("database error")
	}
	delete(m.messages, id)
	return nil
}

func (m *mockMessageRepository) GetAll(ctx context.Context, offset, limit int) ([]*entities.Message, error) {
	if m.shouldFail {
		return nil, errors.New("database error")
	}

	var all []*entities.Message
	for _, msg := range m.messages {
		all = append(all, msg)
	}
	return all, nil
}

type mockCacheRepository struct {
	shouldFail   bool
	sentMessages map[string]mockSentMessage
}

type mockSentMessage struct {
	ExternalID string
	SentAt     time.Time
}

func newMockCacheRepository() *mockCacheRepository {
	return &mockCacheRepository{
		sentMessages: make(map[string]mockSentMessage),
	}
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (m *mockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *mockCacheRepository) SetMessageSent(ctx context.Context, messageID, externalMessageID string, sentAt time.Time) error {
	if m.shouldFail {
		return errors.New("cache error")
	}
	m.sentMessages[messageID] = mockSentMessage{
		ExternalID: externalMessageID,
		SentAt:     sentAt,
	}
	return nil
}

func (m *mockCacheRepository) GetMessageSent(ctx context.Context, messageID string) (string, time.Time, error) {
	if m.shouldFail {
		return "", time.Time{}, errors.New("cache error")
	}
	if msg, exists := m.sentMessages[messageID]; exists {
		return msg.ExternalID, msg.SentAt, nil
	}
	return "", time.Time{}, errors.New("not found")
}

func (m *mockCacheRepository) SetSchedulerStatus(ctx context.Context, status string) error {
	return nil
}

func (m *mockCacheRepository) GetSchedulerStatus(ctx context.Context) (string, error) {
	return "", nil
}

type mockAPIWrapper struct {
	shouldFail  bool
	response    *external.SendMessageResponse
	callCount   int
	lastRequest mockAPIRequest
	sendFunc    func(ctx context.Context, phoneNumber, message string) (*external.SendMessageResponse, error)
}

type mockAPIRequest struct {
	PhoneNumber string
	Message     string
}

func newMockAPIClient() *mockAPIClient {
	return &mockAPIClient{
		response: &external.SendMessageResponse{
			MessageID: "ext_123",
			Status:    "sent",
			Message:   "Success",
		},
	}
}

func (m *mockAPIClient) SendMessage(ctx context.Context, phoneNumber, message string) (*external.SendMessageResponse, error) {
	m.callCount++
	m.lastRequest = mockAPIRequest{
		PhoneNumber: phoneNumber,
		Message:     message,
	}

	if m.sendFunc != nil {
		return m.sendFunc(ctx, phoneNumber, message)
	}

	if m.shouldFail {
		return nil, errors.New("API error")
	}

	return m.response, nil
}

func TestMessageUseCase_CreateMessage(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		phoneNumber string
		repoFail    bool
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "successful creation",
			content:     "Test message",
			phoneNumber: "+1234567890",
			repoFail:    false,
			wantErr:     false,
		},
		{
			name:        "empty content",
			content:     "",
			phoneNumber: "+1234567890",
			repoFail:    false,
			wantErr:     true,
			expectedErr: entities.ErrInvalidMessageContent,
		},
		{
			name:        "empty phone number",
			content:     "Test message",
			phoneNumber: "",
			repoFail:    false,
			wantErr:     true,
			expectedErr: entities.ErrInvalidPhoneNumber,
		},
		{
			name:        "message too long",
			content:     string(make([]byte, 161)),
			phoneNumber: "+1234567890",
			repoFail:    false,
			wantErr:     true,
			expectedErr: entities.ErrMessageTooLong,
		},
		{
			name:        "repository error",
			content:     "Test message",
			phoneNumber: "+1234567890",
			repoFail:    true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockMessageRepository()
			mockRepo.shouldFail = tt.repoFail
			mockCache := newMockCacheRepository()
			mockAPI := newMockAPIClient()
			logger, _ := zap.NewNop(), zap.NewNop()

			useCase := NewMessageUseCase(mockRepo, mockCache, (*external.MessageAPIClient)(mockAPI), logger)

			ctx := context.Background()
			result, err := useCase.CreateMessage(ctx, tt.content, tt.phoneNumber)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result but got nil")
				return
			}

			if result.Content != tt.content {
				t.Errorf("Expected content %v, got %v", tt.content, result.Content)
			}

			if result.PhoneNumber != tt.phoneNumber {
				t.Errorf("Expected phone number %v, got %v", tt.phoneNumber, result.PhoneNumber)
			}

			if result.Status != entities.MessageStatusPending {
				t.Errorf("Expected status %v, got %v", entities.MessageStatusPending, result.Status)
			}
		})
	}
}

func TestMessageUseCase_SendMessage(t *testing.T) {
	tests := []struct {
		name           string
		message        *entities.Message
		apiShouldFail  bool
		cacheFail      bool
		wantErr        bool
		expectedStatus entities.MessageStatus
	}{
		{
			name: "successful send",
			message: &entities.Message{
				ID:          uuid.New(),
				Content:     "Test message",
				PhoneNumber: "+1234567890",
				Status:      entities.MessageStatusPending,
			},
			apiShouldFail:  false,
			cacheFail:      false,
			wantErr:        false,
			expectedStatus: entities.MessageStatusSent,
		},
		{
			name: "API failure",
			message: &entities.Message{
				ID:          uuid.New(),
				Content:     "Test message",
				PhoneNumber: "+1234567890",
				Status:      entities.MessageStatusPending,
			},
			apiShouldFail:  true,
			cacheFail:      false,
			wantErr:        true,
			expectedStatus: entities.MessageStatusFailed,
		},
		{
			name: "cache failure should not fail send",
			message: &entities.Message{
				ID:          uuid.New(),
				Content:     "Test message",
				PhoneNumber: "+1234567890",
				Status:      entities.MessageStatusPending,
			},
			apiShouldFail:  false,
			cacheFail:      true,
			wantErr:        false,
			expectedStatus: entities.MessageStatusSent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockMessageRepository()
			mockCache := newMockCacheRepository()
			mockCache.shouldFail = tt.cacheFail
			mockAPI := newMockAPIClient()
			mockAPI.shouldFail = tt.apiShouldFail
			logger, _ := zap.NewNop(), zap.NewNop()

			useCase := NewMessageUseCase(mockRepo, mockCache, (*external.MessageAPIClient)(mockAPI), logger)

			ctx := context.Background()
			err := useCase.SendMessage(ctx, tt.message)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.message.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, tt.message.Status)
			}

			if tt.expectedStatus == entities.MessageStatusSent {
				if tt.message.SentAt == nil {
					t.Error("Expected SentAt to be set")
				}
				if tt.message.ExternalMessageID == nil {
					t.Error("Expected ExternalMessageID to be set")
				}
			}
		})
	}
}

func TestMessageUseCase_ProcessPendingMessages(t *testing.T) {
	tests := []struct {
		name         string
		batchSize    int
		pendingCount int
		apiFailCount int
		expectedSent int
	}{
		{
			name:         "process all pending successfully",
			batchSize:    5,
			pendingCount: 3,
			apiFailCount: 0,
			expectedSent: 3,
		},
		{
			name:         "batch size limits processing",
			batchSize:    2,
			pendingCount: 5,
			apiFailCount: 0,
			expectedSent: 2,
		},
		{
			name:         "some API failures",
			batchSize:    5,
			pendingCount: 3,
			apiFailCount: 1,
			expectedSent: 2,
		},
		{
			name:         "no pending messages",
			batchSize:    5,
			pendingCount: 0,
			apiFailCount: 0,
			expectedSent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockMessageRepository()
			mockCache := newMockCacheRepository()
			mockAPI := newMockAPIClient()
			logger, _ := zap.NewNop(), zap.NewNop()

			// Create pending messages
			for i := 0; i < tt.pendingCount; i++ {
				msg := &entities.Message{
					ID:          uuid.New(),
					Content:     "Test message",
					PhoneNumber: "+1234567890",
					Status:      entities.MessageStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				mockRepo.Create(context.Background(), msg)
			}

			apiCallCount := 0
			originalSendMessage := mockAPI.SendMessage
			mockAPI.SendMessage = func(ctx context.Context, phoneNumber, message string) (*external.SendMessageResponse, error) {
				apiCallCount++
				if apiCallCount <= tt.apiFailCount {
					return nil, errors.New("API error")
				}
				return originalSendMessage(ctx, phoneNumber, message)
			}

			useCase := NewMessageUseCase(mockRepo, mockCache, (*external.MessageAPIClient)(mockAPI), logger)

			ctx := context.Background()
			sentCount, err := useCase.ProcessPendingMessages(ctx, tt.batchSize)

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if sentCount != tt.expectedSent {
				t.Errorf("Expected %d sent messages, got %d", tt.expectedSent, sentCount)
			}

			expectedAPICalls := min(tt.pendingCount, tt.batchSize)
			if apiCallCount != expectedAPICalls {
				t.Errorf("Expected %d API calls, got %d", expectedAPICalls, apiCallCount)
			}
		})
	}
}

func TestMessageUseCase_GetMessageStats(t *testing.T) {
	mockRepo := newMockMessageRepository()
	mockCache := newMockCacheRepository()
	mockAPI := newMockAPIClient()
	logger, _ := zap.NewNop(), zap.NewNop()

	messages := []*entities.Message{
		{ID: uuid.New(), Status: entities.MessageStatusPending},
		{ID: uuid.New(), Status: entities.MessageStatusPending},
		{ID: uuid.New(), Status: entities.MessageStatusSent},
		{ID: uuid.New(), Status: entities.MessageStatusFailed},
	}

	for _, msg := range messages {
		mockRepo.Create(context.Background(), msg)
	}

	useCase := NewMessageUseCase(mockRepo, mockCache, mockAPI, logger)

	ctx := context.Background()
	stats, err := useCase.GetMessageStats(ctx)

	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if stats.TotalMessages != 4 {
		t.Errorf("Expected 4 total messages, got %d", stats.TotalMessages)
	}

	if stats.PendingMessages != 2 {
		t.Errorf("Expected 2 pending messages, got %d", stats.PendingMessages)
	}

	if stats.SentMessages != 1 {
		t.Errorf("Expected 1 sent message, got %d", stats.SentMessages)
	}

	if stats.FailedMessages != 1 {
		t.Errorf("Expected 1 failed message, got %d", stats.FailedMessages)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
