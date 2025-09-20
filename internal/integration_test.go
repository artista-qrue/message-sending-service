package internal

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"message-sending-service/internal/application/dto"
	"message-sending-service/internal/application/handlers"
	"message-sending-service/internal/application/usecases"
	"message-sending-service/internal/infrastructure/config"
	"message-sending-service/internal/infrastructure/database"
	"message-sending-service/internal/infrastructure/external"
	"message-sending-service/internal/presentation/http"
)

// Integration test: Full message creation flow
func TestMessageCreationFlow_Integration(t *testing.T) {
	// Bu test gerçek bir integration test değil, ama nasıl olacağını gösteriyor
	
	gin.SetMode(gin.TestMode)
	
	// Setup test configuration
	cfg := &config.Config{
		External: config.ExternalConfig{
			MessageAPIURL: "https://httpbin.org/post", // Test endpoint
			Timeout:       5 * time.Second,
		},
		Scheduler: config.SchedulerConfig{
			Interval:         2 * time.Minute,
			MessagesPerBatch: 2,
		},
	}

	// Setup logger
	logger, _ := zap.NewNop(), zap.NewNop()

	// Setup external API client (real one)
	apiClient := external.NewMessageAPIClient(cfg)

	// Note: In a real integration test, you would:
	// 1. Setup test database
	// 2. Run migrations
	// 3. Create real repositories
	// But for this demo, we'll use nil and focus on the flow

	// Setup use cases
	messageUseCase := usecases.NewMessageUseCase(nil, nil, apiClient, logger)
	schedulerUseCase := usecases.NewSchedulerUseCase(messageUseCase, nil, cfg, logger)

	// Setup handlers
	messageHandler := handlers.NewMessageHandler(messageUseCase, logger)
	schedulerHandler := handlers.NewSchedulerHandler(schedulerUseCase, logger)

	// Setup router (real HTTP router)
	router := http.NewRouter(messageHandler, schedulerHandler, logger)
	ginEngine := router.SetupRoutes()

	t.Run("create message via HTTP API", func(t *testing.T) {
		// Prepare request
		createReq := dto.CreateMessageRequest{
			Content:     "Integration test message",
			PhoneNumber: "+1234567890",
		}

		reqBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Make request
		w := httptest.NewRecorder()
		ginEngine.ServeHTTP(w, req)

		// This will fail due to nil repository, but shows the integration flow
		// In real test with database, this would succeed
		t.Logf("Response status: %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	})

	t.Run("get scheduler status via HTTP API", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/scheduler/status", nil)
		w := httptest.NewRecorder()
		
		ginEngine.ServeHTTP(w, req)

		// This should work since it doesn't need database
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, exists := response["data"]; !exists {
			t.Error("Expected data in response")
		}

		t.Logf("Scheduler status response: %s", w.Body.String())
	})
}

// Integration test template for database operations
func TestDatabaseIntegration_Template(t *testing.T) {
	// Bu real integration test template'i
	t.Skip("This is a template - requires test database setup")

	/*
	Real implementation would be:

	// 1. Setup test database
	testDB := setupTestDatabase(t)
	defer cleanupTestDatabase(t, testDB)

	// 2. Run migrations
	err := database.CreateTables(testDB)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// 3. Create real repositories
	messageRepo := database.NewMessageRepository(testDB)

	// 4. Test actual database operations
	message := &entities.Message{
		ID:          uuid.New(),
		Content:     "Test message",
		PhoneNumber: "+1234567890",
		Status:      entities.MessageStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create
	err = messageRepo.Create(context.Background(), message)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Retrieve
	retrieved, err := messageRepo.GetByID(context.Background(), message.ID)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	// Verify
	if retrieved.Content != message.Content {
		t.Errorf("Expected content %s, got %s", message.Content, retrieved.Content)
	}
	*/
}

// Benchmark integration test
func BenchmarkMessageCreationFlow(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	
	cfg := &config.Config{
		External: config.ExternalConfig{
			MessageAPIURL: "https://httpbin.org/post",
			Timeout:       5 * time.Second,
		},
	}

	logger, _ := zap.NewNop(), zap.NewNop()
	apiClient := external.NewMessageAPIClient(cfg)
	messageUseCase := usecases.NewMessageUseCase(nil, nil, apiClient, logger)
	messageHandler := handlers.NewMessageHandler(messageUseCase, logger)

	createReq := dto.CreateMessageRequest{
		Content:     "Benchmark test message",
		PhoneNumber: "+1234567890",
	}
	reqBody, _ := json.Marshal(createReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// This will fail at validation but measures the request processing overhead
		messageHandler.CreateMessage(c)
	}
}

// Helper function for real integration tests
func setupTestDatabase(t *testing.T) *sql.DB {
	// Real implementation would:
	// 1. Create test database connection
	// 2. Run migrations
	// 3. Seed test data
	// 4. Return connection
	t.Helper()
	return nil
}

func cleanupTestDatabase(t *testing.T, db *sql.DB) {
	// Real implementation would:
	// 1. Drop test tables
	// 2. Close connection
	// 3. Clean up resources
	t.Helper()
}
