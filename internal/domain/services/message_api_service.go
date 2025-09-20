package services

import (
	"context"

	"message-sending-service/internal/infrastructure/external"
)

type MessageAPIService interface {
	SendMessage(ctx context.Context, phoneNumber, message string) (*external.SendMessageResponse, error)
}
