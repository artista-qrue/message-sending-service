package repositories

import (
	"context"

	"github.com/google/uuid"
	"message-sending-service/internal/domain/entities"
)

type MessageRepository interface {
	Create(ctx context.Context, message *entities.Message) error

	GetByID(ctx context.Context, id uuid.UUID) (*entities.Message, error)

	GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error)

	GetSentMessages(ctx context.Context, offset, limit int) ([]*entities.Message, error)

	Update(ctx context.Context, message *entities.Message) error

	Delete(ctx context.Context, id uuid.UUID) error

	CountByStatus(ctx context.Context, status entities.MessageStatus) (int64, error)

	GetAll(ctx context.Context, offset, limit int) ([]*entities.Message, error)
}
