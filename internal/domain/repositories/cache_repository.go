package repositories

import (
	"context"
	"time"
)

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	Get(ctx context.Context, key string) (string, error)

	Delete(ctx context.Context, key string) error

	Exists(ctx context.Context, key string) (bool, error)

	SetMessageSent(ctx context.Context, messageID, externalMessageID string, sentAt time.Time) error

	GetMessageSent(ctx context.Context, messageID string) (string, time.Time, error)

	SetSchedulerStatus(ctx context.Context, status string) error

	GetSchedulerStatus(ctx context.Context) (string, error)
}
