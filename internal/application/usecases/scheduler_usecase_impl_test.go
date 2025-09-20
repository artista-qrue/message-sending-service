package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"message-sending-service/internal/domain/entities"
	domainUsecases "message-sending-service/internal/domain/usecases"
	"message-sending-service/internal/infrastructure/config"
)

type mockMessageUseCase struct {
	processedCount int
	shouldFail     bool
	callCount      int
}

func newMockMessageUseCase() *mockMessageUseCase {
	return &mockMessageUseCase{}
}

func (m *mockMessageUseCase) CreateMessage(ctx context.Context, content, phoneNumber string) (*entities.Message, error) {
	return nil, nil
}

func (m *mockMessageUseCase) GetMessageByID(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
	return nil, nil
}

func (m *mockMessageUseCase) GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error) {
	return nil, nil
}

func (m *mockMessageUseCase) GetSentMessages(ctx context.Context, page, limit int) ([]*entities.Message, int64, error) {
	return nil, 0, nil
}

func (m *mockMessageUseCase) SendMessage(ctx context.Context, message *entities.Message) error {
	return nil
}

func (m *mockMessageUseCase) ProcessPendingMessages(ctx context.Context, batchSize int) (int, error) {
	m.callCount++
	if m.shouldFail {
		return 0, entities.ErrMessageNotFound
	}
	return m.processedCount, nil
}

func (m *mockMessageUseCase) GetMessageStats(ctx context.Context) (*domainUsecases.MessageStats, error) {
	return nil, nil
}

func TestSchedulerUseCase_StartScheduler(t *testing.T) {
	tests := []struct {
		name           string
		alreadyRunning bool
		wantErr        bool
		expectedErr    error
	}{
		{
			name:           "start scheduler successfully",
			alreadyRunning: false,
			wantErr:        false,
		},
		{
			name:           "scheduler already running",
			alreadyRunning: true,
			wantErr:        true,
			expectedErr:    entities.ErrSchedulerAlreadyRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMessageUC := newMockMessageUseCase()
			mockCache := newMockCacheRepository()
			cfg := &config.Config{
				Scheduler: config.SchedulerConfig{
					Interval:         1 * time.Second,
					MessagesPerBatch: 2,
				},
			}
			logger, _ := zap.NewNop(), zap.NewNop()

			useCase := NewSchedulerUseCase(mockMessageUC, mockCache, cfg, logger)

			if tt.alreadyRunning {
				err := useCase.StartScheduler(context.Background())
				if err != nil {
					t.Fatalf("Failed to start scheduler for test setup: %v", err)
				}
			}

			ctx := context.Background()
			err := useCase.StartScheduler(ctx)

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

				status, err := useCase.GetSchedulerStatus(ctx)
				if err != nil {
					t.Errorf("Failed to get scheduler status: %v", err)
				}
				if status.Status != entities.SchedulerStatusRunning {
					t.Errorf("Expected scheduler to be running, got %v", status.Status)
				}

				useCase.StopScheduler(ctx)
			}
		})
	}
}

func TestSchedulerUseCase_StopScheduler(t *testing.T) {
	tests := []struct {
		name        string
		isRunning   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "stop running scheduler",
			isRunning: true,
			wantErr:   false,
		},
		{
			name:        "stop already stopped scheduler",
			isRunning:   false,
			wantErr:     true,
			expectedErr: entities.ErrSchedulerNotRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMessageUC := newMockMessageUseCase()
			mockCache := newMockCacheRepository()
			cfg := &config.Config{
				Scheduler: config.SchedulerConfig{
					Interval:         1 * time.Second,
					MessagesPerBatch: 2,
				},
			}
			logger, _ := zap.NewNop(), zap.NewNop()

			useCase := NewSchedulerUseCase(mockMessageUC, mockCache, cfg, logger)

			ctx := context.Background()

			if tt.isRunning {
				err := useCase.StartScheduler(ctx)
				if err != nil {
					t.Fatalf("Failed to start scheduler for test setup: %v", err)
				}
			}

			err := useCase.StopScheduler(ctx)

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

				status, err := useCase.GetSchedulerStatus(ctx)
				if err != nil {
					t.Errorf("Failed to get scheduler status: %v", err)
				}
				if status.Status != entities.SchedulerStatusStopped {
					t.Errorf("Expected scheduler to be stopped, got %v", status.Status)
				}
			}
		})
	}
}

func TestSchedulerUseCase_GetSchedulerStatus(t *testing.T) {
	mockMessageUC := newMockMessageUseCase()
	mockCache := newMockCacheRepository()
	cfg := &config.Config{
		Scheduler: config.SchedulerConfig{
			Interval:         2 * time.Minute,
			MessagesPerBatch: 2,
		},
	}
	logger, _ := zap.NewNop(), zap.NewNop()

	useCase := NewSchedulerUseCase(mockMessageUC, mockCache, cfg, logger)

	ctx := context.Background()

	status, err := useCase.GetSchedulerStatus(ctx)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if status.Status != entities.SchedulerStatusStopped {
		t.Errorf("Expected status %v, got %v", entities.SchedulerStatusStopped, status.Status)
	}

	if status.Interval != 2*time.Minute {
		t.Errorf("Expected interval %v, got %v", 2*time.Minute, status.Interval)
	}

	if status.BatchSize != 2 {
		t.Errorf("Expected batch size %d, got %d", 2, status.BatchSize)
	}

	err = useCase.StartScheduler(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer useCase.StopScheduler(ctx)

	status, err = useCase.GetSchedulerStatus(ctx)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if status.Status != entities.SchedulerStatusRunning {
		t.Errorf("Expected status %v, got %v", entities.SchedulerStatusRunning, status.Status)
	}

	if status.NextRun == nil {
		t.Error("Expected NextRun to be set when scheduler is running")
	}
}

func TestSchedulerUseCase_IsRunning(t *testing.T) {
	mockMessageUC := newMockMessageUseCase()
	mockCache := newMockCacheRepository()
	cfg := &config.Config{
		Scheduler: config.SchedulerConfig{
			Interval:         1 * time.Second,
			MessagesPerBatch: 2,
		},
	}
	logger, _ := zap.NewNop(), zap.NewNop()

	useCase := NewSchedulerUseCase(mockMessageUC, mockCache, cfg, logger)

	ctx := context.Background()

	if useCase.IsRunning(ctx) {
		t.Error("Expected scheduler to not be running initially")
	}

	err := useCase.StartScheduler(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	if !useCase.IsRunning(ctx) {
		t.Error("Expected scheduler to be running after start")
	}

	err = useCase.StopScheduler(ctx)
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}

	if useCase.IsRunning(ctx) {
		t.Error("Expected scheduler to not be running after stop")
	}
}

func TestSchedulerUseCase_ProcessMessages_Integration(t *testing.T) {
	mockMessageUC := newMockMessageUseCase()
	mockMessageUC.processedCount = 2
	mockCache := newMockCacheRepository()
	cfg := &config.Config{
		Scheduler: config.SchedulerConfig{
			Interval:         100 * time.Millisecond,
			MessagesPerBatch: 2,
		},
	}
	logger, _ := zap.NewNop(), zap.NewNop()

	useCase := NewSchedulerUseCase(mockMessageUC, mockCache, cfg, logger)

	ctx := context.Background()

	err := useCase.StartScheduler(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer useCase.StopScheduler(ctx)

	time.Sleep(200 * time.Millisecond)

	if mockMessageUC.callCount == 0 {
		t.Error("Expected ProcessPendingMessages to be called but it wasn't")
	}

	status, err := useCase.GetSchedulerStatus(ctx)
	if err != nil {
		t.Errorf("Failed to get scheduler status: %v", err)
	}

	if status.LastRun == nil {
		t.Error("Expected LastRun to be set after processing")
	}

	expectedCount := mockMessageUC.callCount * mockMessageUC.processedCount
	if status.MessagesCount != expectedCount {
		t.Errorf("Expected message count %d, got %d", expectedCount, status.MessagesCount)
	}
}

func TestSchedulerUseCase_CacheIntegration(t *testing.T) {
	mockMessageUC := newMockMessageUseCase()
	mockCache := newMockCacheRepository()
	cfg := &config.Config{
		Scheduler: config.SchedulerConfig{
			Interval:         1 * time.Second,
			MessagesPerBatch: 2,
		},
	}
	logger, _ := zap.NewNop(), zap.NewNop()

	useCase := NewSchedulerUseCase(mockMessageUC, mockCache, cfg, logger)

	ctx := context.Background()

	err := useCase.StartScheduler(ctx)
	if err != nil {
		t.Errorf("Failed to start scheduler: %v", err)
	}

	mockCache.shouldFail = true

	err = useCase.StopScheduler(ctx)
	if err != nil {
		t.Errorf("Stop scheduler should not fail even if cache fails: %v", err)
	}
}
