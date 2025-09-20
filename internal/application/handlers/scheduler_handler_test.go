package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"message-sending-service/internal/domain/entities"
)

// Mock scheduler testleri
type mockSchedulerUseCase struct {
	startSchedulerFunc func(ctx context.Context) error
	stopSchedulerFunc  func(ctx context.Context) error
	getStatusFunc      func(ctx context.Context) (*entities.SchedulerInfo, error)
	isRunningFunc      func(ctx context.Context) bool
}

func (m *mockSchedulerUseCase) StartScheduler(ctx context.Context) error {
	if m.startSchedulerFunc != nil {
		return m.startSchedulerFunc(ctx)
	}
	return nil
}

func (m *mockSchedulerUseCase) StopScheduler(ctx context.Context) error {
	if m.stopSchedulerFunc != nil {
		return m.stopSchedulerFunc(ctx)
	}
	return nil
}

func (m *mockSchedulerUseCase) GetSchedulerStatus(ctx context.Context) (*entities.SchedulerInfo, error) {
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx)
	}
	now := time.Now()
	return &entities.SchedulerInfo{
		Status:        entities.SchedulerStatusRunning,
		LastRun:       &now,
		NextRun:       &now,
		MessagesCount: 10,
		Interval:      2 * time.Minute,
		BatchSize:     2,
	}, nil
}

func (m *mockSchedulerUseCase) IsRunning(ctx context.Context) bool {
	if m.isRunningFunc != nil {
		return m.isRunningFunc(ctx)
	}
	return true
}

func TestSchedulerHandler_StartScheduler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockFunc       func(ctx context.Context) error
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful start",
			mockFunc:       nil, // Use default mock (success)
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "scheduler already running",
			mockFunc: func(ctx context.Context) error {
				return entities.ErrSchedulerAlreadyRunning
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "internal error",
			mockFunc: func(ctx context.Context) error {
				return entities.ErrMessageNotFound // Any other error
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mockSchedulerUseCase{
				startSchedulerFunc: tt.mockFunc,
			}
			logger, _ := zap.NewNop(), zap.NewNop()
			handler := NewSchedulerHandler(mockUseCase, logger)

			req := httptest.NewRequest("POST", "/scheduler/start", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.StartScheduler(c)

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

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Expected data to be an object")
				} else {
					if status, exists := data["status"]; !exists || status != "running" {
						t.Error("Expected status to be 'running' in successful start response")
					}
				}
			}
		})
	}
}

func TestSchedulerHandler_StopScheduler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockFunc       func(ctx context.Context) error
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful stop",
			mockFunc:       nil,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "scheduler not running",
			mockFunc: func(ctx context.Context) error {
				return entities.ErrSchedulerNotRunning
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "internal error",
			mockFunc: func(ctx context.Context) error {
				return entities.ErrMessageNotFound // Any other error
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mockSchedulerUseCase{
				stopSchedulerFunc: tt.mockFunc,
			}
			logger, _ := zap.NewNop(), zap.NewNop()
			handler := NewSchedulerHandler(mockUseCase, logger)

			req := httptest.NewRequest("POST", "/scheduler/stop", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.StopScheduler(c)

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

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Expected data to be an object")
				} else {
					if status, exists := data["status"]; !exists || status != "stopped" {
						t.Error("Expected status to be 'stopped' in successful stop response")
					}
				}
			}
		})
	}
}

func TestSchedulerHandler_GetSchedulerStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockFunc       func(ctx context.Context) (*entities.SchedulerInfo, error)
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful get status",
			mockFunc:       nil, // Use default mock
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "get status error",
			mockFunc: func(ctx context.Context) (*entities.SchedulerInfo, error) {
				return nil, entities.ErrMessageNotFound // Any error
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mockSchedulerUseCase{
				getStatusFunc: tt.mockFunc,
			}
			logger, _ := zap.NewNop(), zap.NewNop()
			handler := NewSchedulerHandler(mockUseCase, logger)

			req := httptest.NewRequest("GET", "/scheduler/status", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.GetSchedulerStatus(c)

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

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Expected data to be an object")
				} else {
					expectedFields := []string{"status", "messages_sent_count", "interval", "batch_size"}
					for _, field := range expectedFields {
						if _, exists := data[field]; !exists {
							t.Errorf("Expected field %s in scheduler status data", field)
						}
					}
				}
			}
		})
	}
}

func TestSchedulerHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("start scheduler with context cancellation", func(t *testing.T) {
		mockUseCase := &mockSchedulerUseCase{
			startSchedulerFunc: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return nil
				}
			},
		}
		logger, _ := zap.NewNop(), zap.NewNop()
		handler := NewSchedulerHandler(mockUseCase, logger)

		req := httptest.NewRequest("POST", "/scheduler/start", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.StartScheduler(c)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("get status with different scheduler states", func(t *testing.T) {
		states := []entities.SchedulerStatus{
			entities.SchedulerStatusRunning,
			entities.SchedulerStatusStopped,
		}

		for _, state := range states {
			mockUseCase := &mockSchedulerUseCase{
				getStatusFunc: func(ctx context.Context) (*entities.SchedulerInfo, error) {
					return &entities.SchedulerInfo{
						Status:        state,
						MessagesCount: 5,
						Interval:      2 * time.Minute,
						BatchSize:     2,
					}, nil
				},
			}
			logger, _ := zap.NewNop(), zap.NewNop()
			handler := NewSchedulerHandler(mockUseCase, logger)

			req := httptest.NewRequest("GET", "/scheduler/status", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.GetSchedulerStatus(c)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d for state %s, got %d", http.StatusOK, state, w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})

			if data["status"] != string(state) {
				t.Errorf("Expected status %s, got %s", state, data["status"])
			}
		}
	})
}
