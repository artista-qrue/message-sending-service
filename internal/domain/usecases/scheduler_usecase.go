package usecases

import (
	"context"

	"message-sending-service/internal/domain/entities"
)

type SchedulerUseCase interface {
	StartScheduler(ctx context.Context) error

	StopScheduler(ctx context.Context) error

	GetSchedulerStatus(ctx context.Context) (*entities.SchedulerInfo, error)

	IsRunning(ctx context.Context) bool
}
