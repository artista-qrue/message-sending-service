package external

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"message-sending-service/internal/infrastructure/config"
)

func TestMessageAPIClient_SendMessage(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		phoneNumber    string
		message        string
		expectError    bool
		expectedStatus string
	}{
		{
			name:           "successful send",
			serverResponse: `{"message_id": "test_123", "status": "sent"}`,
			serverStatus:   200,
			phoneNumber:    "+1234567890",
			message:        "Test message",
			expectError:    false,
			expectedStatus: "sent",
		},
		{
			name:           "server error",
			serverResponse: `{"error": "Server error"}`,
			serverStatus:   500,
			phoneNumber:    "+1234567890",
			message:        "Test message",
			expectError:    true,
			expectedStatus: "",
		},
		{
			name:           "success with webhook response",
			serverResponse: `{"success": true}`,
			serverStatus:   200,
			phoneNumber:    "+1234567890",
			message:        "Test message",
			expectError:    false,
			expectedStatus: "sent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			cfg := &config.Config{
				External: config.ExternalConfig{
					MessageAPIURL: server.URL,
					Timeout:       5 * time.Second,
				},
			}
			client := NewMessageAPIClient(cfg)

			ctx := context.Background()
			response, err := client.SendMessage(ctx, tt.phoneNumber, tt.message)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError && response != nil {
				if response.Status != tt.expectedStatus {
					t.Errorf("Expected status %v, got %v", tt.expectedStatus, response.Status)
				}
				if response.MessageID == "" {
					t.Error("Expected message ID to be set")
				}
			}
		})
	}
}

func TestMessageAPIClient_SendMessage_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte(`{"message_id": "test", "status": "sent"}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		External: config.ExternalConfig{
			MessageAPIURL: server.URL,
			Timeout:       50 * time.Millisecond,
		},
	}
	client := NewMessageAPIClient(cfg)

	ctx := context.Background()
	_, err := client.SendMessage(ctx, "+1234567890", "Test")

	if err == nil {
		t.Error("Expected timeout error but got none")
	}
}

func TestGenerateMockMessageID(t *testing.T) {
	id1 := generateMockMessageID()
	id2 := generateMockMessageID()

	if id1 == id2 {
		t.Error("Expected different message IDs")
	}

	if id1 == "" {
		t.Error("Expected non-empty message ID")
	}

	if len(id1) < 10 {
		t.Error("Expected message ID to have reasonable length")
	}
}
