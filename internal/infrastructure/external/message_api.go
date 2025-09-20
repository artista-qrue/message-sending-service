package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"message-sending-service/internal/infrastructure/config"
)

type MessageAPIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMessageAPIClient(cfg *config.Config) *MessageAPIClient {
	return &MessageAPIClient{
		baseURL: cfg.External.MessageAPIURL,
		httpClient: &http.Client{
			Timeout: cfg.External.Timeout,
		},
	}
}

type SendMessageRequest struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
}

type SendMessageResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (c *MessageAPIClient) SendMessage(ctx context.Context, phoneNumber, message string) (*SendMessageResponse, error) {
	request := SendMessageRequest{
		PhoneNumber: phoneNumber,
		Message:     message,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Insider-Sending-Service/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response := &SendMessageResponse{
			MessageID: generateMockMessageID(),
			Status:    "sent",
			Message:   "Message sent successfully",
		}

		var apiResponse SendMessageResponse
		if err := json.Unmarshal(body, &apiResponse); err == nil && apiResponse.MessageID != "" {
			response = &apiResponse
		}

		return response, nil
	}

	var errorResponse SendMessageResponse
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		errorResponse = SendMessageResponse{
			Status: "failed",
			Error:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	return &errorResponse, fmt.Errorf("API request failed with status %d", resp.StatusCode)
}

func generateMockMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
