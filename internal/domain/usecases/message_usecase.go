package usecases

import (
	"context"

	"github.com/google/uuid"
	"message-sending-service/internal/domain/entities"
)

type MessageUseCase interface {
	CreateMessage(ctx context.Context, content, phoneNumber string) (*entities.Message, error)

	GetMessageByID(ctx context.Context, id uuid.UUID) (*entities.Message, error)

	GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error)

	GetSentMessages(ctx context.Context, page, limit int) ([]*entities.Message, int64, error)

	SendMessage(ctx context.Context, message *entities.Message) error

	ProcessPendingMessages(ctx context.Context, batchSize int) (int, error)

	GetMessageStats(ctx context.Context) (*MessageStats, error)
}

type MessageStats struct {
	TotalMessages   int64 `json:"total_messages"`
	PendingMessages int64 `json:"pending_messages"`
	SentMessages    int64 `json:"sent_messages"`
	FailedMessages  int64 `json:"failed_messages"`
}
