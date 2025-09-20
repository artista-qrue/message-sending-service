package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"message-sending-service/internal/domain/entities"
	"message-sending-service/internal/domain/repositories"
	"message-sending-service/internal/domain/usecases"
	"message-sending-service/internal/infrastructure/external"
)

type messageUseCaseImpl struct {
	messageRepo repositories.MessageRepository
	cacheRepo   repositories.CacheRepository
	apiClient   *external.MessageAPIClient
	logger      *zap.Logger
}

func NewMessageUseCase(
	messageRepo repositories.MessageRepository,
	cacheRepo repositories.CacheRepository,
	apiClient *external.MessageAPIClient,
	logger *zap.Logger,
) usecases.MessageUseCase {
	return &messageUseCaseImpl{
		messageRepo: messageRepo,
		cacheRepo:   cacheRepo,
		apiClient:   apiClient,
		logger:      logger,
	}
}

func (uc *messageUseCaseImpl) CreateMessage(ctx context.Context, content, phoneNumber string) (*entities.Message, error) {
	now := time.Now()
	message := &entities.Message{
		ID:          uuid.New(),
		Content:     content,
		PhoneNumber: phoneNumber,
		Status:      entities.MessageStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := message.Validate(); err != nil {
		return nil, err
	}

	if err := uc.messageRepo.Create(ctx, message); err != nil {
		uc.logger.Error("Failed to create message", zap.Error(err))
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	uc.logger.Info("Message created successfully", zap.String("message_id", message.ID.String()))
	return message, nil
}

func (uc *messageUseCaseImpl) GetMessageByID(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
	message, err := uc.messageRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get message by ID", zap.String("message_id", id.String()), zap.Error(err))
		return nil, err
	}

	return message, nil
}

func (uc *messageUseCaseImpl) GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error) {
	messages, err := uc.messageRepo.GetPendingMessages(ctx, limit)
	if err != nil {
		uc.logger.Error("Failed to get pending messages", zap.Error(err))
		return nil, err
	}

	return messages, nil
}

func (uc *messageUseCaseImpl) GetSentMessages(ctx context.Context, page, limit int) ([]*entities.Message, int64, error) {
	offset := (page - 1) * limit

	messages, err := uc.messageRepo.GetSentMessages(ctx, offset, limit)
	if err != nil {
		uc.logger.Error("Failed to get sent messages", zap.Error(err))
		return nil, 0, err
	}

	totalCount, err := uc.messageRepo.CountByStatus(ctx, entities.MessageStatusSent)
	if err != nil {
		uc.logger.Error("Failed to count sent messages", zap.Error(err))
		return nil, 0, err
	}

	return messages, totalCount, nil
}

func (uc *messageUseCaseImpl) SendMessage(ctx context.Context, message *entities.Message) error {
	uc.logger.Info("Sending message",
		zap.String("message_id", message.ID.String()),
		zap.String("phone_number", message.PhoneNumber))

	response, err := uc.apiClient.SendMessage(ctx, message.PhoneNumber, message.Content)
	if err != nil {
		message.MarkAsFailed(err.Error())
		if updateErr := uc.messageRepo.Update(ctx, message); updateErr != nil {
			uc.logger.Error("Failed to update message status after API error",
				zap.String("message_id", message.ID.String()),
				zap.Error(updateErr))
		}

		uc.logger.Error("Failed to send message via API",
			zap.String("message_id", message.ID.String()),
			zap.Error(err))
		return err
	}

	if response.Status != "sent" {
		errorMsg := response.Error
		if errorMsg == "" {
			errorMsg = "Unknown error from API"
		}

		message.MarkAsFailed(errorMsg)
		if updateErr := uc.messageRepo.Update(ctx, message); updateErr != nil {
			uc.logger.Error("Failed to update message status after API failure",
				zap.String("message_id", message.ID.String()),
				zap.Error(updateErr))
		}

		return fmt.Errorf("API returned error: %s", errorMsg)
	}

	message.MarkAsSent(response.MessageID)
	if err := uc.messageRepo.Update(ctx, message); err != nil {
		uc.logger.Error("Failed to update message status after successful send",
			zap.String("message_id", message.ID.String()),
			zap.Error(err))
		return err
	}

	if uc.cacheRepo != nil {
		if err := uc.cacheRepo.SetMessageSent(ctx, message.ID.String(), response.MessageID, *message.SentAt); err != nil {
			uc.logger.Warn("Failed to cache sent message info",
				zap.String("message_id", message.ID.String()),
				zap.Error(err))
		}
	}

	uc.logger.Info("Message sent successfully",
		zap.String("message_id", message.ID.String()),
		zap.String("external_message_id", response.MessageID))

	return nil
}

func (uc *messageUseCaseImpl) ProcessPendingMessages(ctx context.Context, batchSize int) (int, error) {
	uc.logger.Info("Processing pending messages", zap.Int("batch_size", batchSize))

	messages, err := uc.GetPendingMessages(ctx, batchSize)
	if err != nil {
		return 0, err
	}

	if len(messages) == 0 {
		uc.logger.Debug("No pending messages to process")
		return 0, nil
	}

	successCount := 0
	for _, message := range messages {
		if err := uc.SendMessage(ctx, message); err != nil {
			uc.logger.Error("Failed to send message",
				zap.String("message_id", message.ID.String()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	uc.logger.Info("Batch processing completed",
		zap.Int("total_messages", len(messages)),
		zap.Int("successful_sends", successCount),
		zap.Int("failed_sends", len(messages)-successCount))

	return successCount, nil
}

func (uc *messageUseCaseImpl) GetMessageStats(ctx context.Context) (*usecases.MessageStats, error) {
	pendingCount, err := uc.messageRepo.CountByStatus(ctx, entities.MessageStatusPending)
	if err != nil {
		return nil, err
	}

	sentCount, err := uc.messageRepo.CountByStatus(ctx, entities.MessageStatusSent)
	if err != nil {
		return nil, err
	}

	failedCount, err := uc.messageRepo.CountByStatus(ctx, entities.MessageStatusFailed)
	if err != nil {
		return nil, err
	}

	totalCount := pendingCount + sentCount + failedCount

	return &usecases.MessageStats{
		TotalMessages:   totalCount,
		PendingMessages: pendingCount,
		SentMessages:    sentCount,
		FailedMessages:  failedCount,
	}, nil
}
