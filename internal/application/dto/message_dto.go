package dto

import (
	"time"

	"github.com/google/uuid"
	"message-sending-service/internal/domain/entities"
	"message-sending-service/internal/domain/usecases"
)

type CreateMessageRequest struct {
	Content     string `json:"content" binding:"required,max=160" example:"Hello, this is a test message"`
	PhoneNumber string `json:"phone_number" binding:"required" example:"+1234567890"`
}
type MessageResponse struct {
	ID                uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Content           string     `json:"content" example:"Hello, this is a test message"`
	PhoneNumber       string     `json:"phone_number" example:"+1234567890"`
	Status            string     `json:"status" example:"sent"`
	CreatedAt         time.Time  `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt         time.Time  `json:"updated_at" example:"2023-01-01T12:05:00Z"`
	SentAt            *time.Time `json:"sent_at,omitempty" example:"2023-01-01T12:05:00Z"`
	ExternalMessageID *string    `json:"external_message_id,omitempty" example:"ext_msg_123"`
	ErrorMessage      *string    `json:"error_message,omitempty" example:"Network error"`
}

type GetSentMessagesResponse struct {
	Messages   []MessageResponse `json:"messages"`
	TotalCount int64             `json:"total_count" example:"100"`
	Page       int               `json:"page" example:"1"`
	Limit      int               `json:"limit" example:"10"`
	TotalPages int               `json:"total_pages" example:"10"`
}

type MessageStatsResponse struct {
	TotalMessages   int64 `json:"total_messages" example:"1000"`
	PendingMessages int64 `json:"pending_messages" example:"50"`
	SentMessages    int64 `json:"sent_messages" example:"900"`
	FailedMessages  int64 `json:"failed_messages" example:"50"`
}

func ToMessageResponse(message *entities.Message) MessageResponse {
	return MessageResponse{
		ID:                message.ID,
		Content:           message.Content,
		PhoneNumber:       message.PhoneNumber,
		Status:            string(message.Status),
		CreatedAt:         message.CreatedAt,
		UpdatedAt:         message.UpdatedAt,
		SentAt:            message.SentAt,
		ExternalMessageID: message.ExternalMessageID,
		ErrorMessage:      message.ErrorMessage,
	}
}

func ToMessageStatsResponse(stats *usecases.MessageStats) MessageStatsResponse {
	return MessageStatsResponse{
		TotalMessages:   stats.TotalMessages,
		PendingMessages: stats.PendingMessages,
		SentMessages:    stats.SentMessages,
		FailedMessages:  stats.FailedMessages,
	}
}
