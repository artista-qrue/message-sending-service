package usecases

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"message-sending-service/internal/domain/entities"
	"message-sending-service/internal/domain/repositories"
	"message-sending-service/internal/domain/usecases"
	"message-sending-service/internal/infrastructure/config"
)

type schedulerUseCaseImpl struct {
	messageUseCase usecases.MessageUseCase
	cacheRepo      repositories.CacheRepository
	config         *config.Config
	logger         *zap.Logger

	cron      *cron.Cron
	entryID   cron.EntryID
	mu        sync.RWMutex
	isRunning bool
	lastRun   *time.Time
	nextRun   *time.Time
	sentCount int
}

func NewSchedulerUseCase(
	messageUseCase usecases.MessageUseCase,
	cacheRepo repositories.CacheRepository,
	config *config.Config,
	logger *zap.Logger,
) usecases.SchedulerUseCase {
	return &schedulerUseCaseImpl{
		messageUseCase: messageUseCase,
		cacheRepo:      cacheRepo,
		config:         config,
		logger:         logger,
		cron:           cron.New(),
		isRunning:      false,
		sentCount:      0,
	}
}

func (uc *schedulerUseCaseImpl) StartScheduler(ctx context.Context) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.isRunning {
		return entities.ErrSchedulerAlreadyRunning
	}

	cronExpr := "@every " + uc.config.Scheduler.Interval.String()

	entryID, err := uc.cron.AddFunc(cronExpr, func() {
		uc.processMessages()
	})
	if err != nil {
		uc.logger.Error("Failed to add cron job", zap.Error(err))
		return err
	}

	uc.entryID = entryID
	uc.cron.Start()
	uc.isRunning = true

	now := time.Now()
	uc.nextRun = &now

	if uc.cacheRepo != nil {
		if err := uc.cacheRepo.SetSchedulerStatus(ctx, string(entities.SchedulerStatusRunning)); err != nil {
			uc.logger.Warn("Failed to cache scheduler status", zap.Error(err))
		}
	}

	uc.logger.Info("Scheduler started",
		zap.String("interval", uc.config.Scheduler.Interval.String()),
		zap.Int("batch_size", uc.config.Scheduler.MessagesPerBatch))

	return nil
}

func (uc *schedulerUseCaseImpl) StopScheduler(ctx context.Context) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if !uc.isRunning {
		return entities.ErrSchedulerNotRunning
	}

	uc.cron.Remove(uc.entryID)
	uc.cron.Stop()
	uc.isRunning = false
	uc.nextRun = nil

	if uc.cacheRepo != nil {
		if err := uc.cacheRepo.SetSchedulerStatus(ctx, string(entities.SchedulerStatusStopped)); err != nil {
			uc.logger.Warn("Failed to cache scheduler status", zap.Error(err))
		}
	}

	uc.logger.Info("Scheduler stopped")
	return nil
}

func (uc *schedulerUseCaseImpl) GetSchedulerStatus(ctx context.Context) (*entities.SchedulerInfo, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	status := entities.SchedulerStatusStopped
	if uc.isRunning {
		status = entities.SchedulerStatusRunning
	}

	info := &entities.SchedulerInfo{
		Status:        status,
		LastRun:       uc.lastRun,
		NextRun:       uc.nextRun,
		MessagesCount: uc.sentCount,
		Interval:      uc.config.Scheduler.Interval,
		BatchSize:     uc.config.Scheduler.MessagesPerBatch,
	}

	return info, nil
}

func (uc *schedulerUseCaseImpl) IsRunning(ctx context.Context) bool {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	return uc.isRunning
}

func (uc *schedulerUseCaseImpl) processMessages() {
	ctx := context.Background()

	uc.mu.Lock()
	now := time.Now()
	uc.lastRun = &now

	nextRun := now.Add(uc.config.Scheduler.Interval)
	uc.nextRun = &nextRun
	uc.mu.Unlock()

	uc.logger.Info("Starting scheduled message processing")

	sentCount, err := uc.messageUseCase.ProcessPendingMessages(ctx, uc.config.Scheduler.MessagesPerBatch)
	if err != nil {
		uc.logger.Error("Failed to process pending messages", zap.Error(err))
		return
	}

	uc.mu.Lock()
	uc.sentCount += sentCount
	uc.mu.Unlock()

	uc.logger.Info("Scheduled message processing completed",
		zap.Int("messages_sent", sentCount),
		zap.Int("total_sent", uc.sentCount))
}
