package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"message-sending-service/internal/domain/repositories"
)

type cacheRepositoryImpl struct {
	client *redis.Client
}

func NewCacheRepository(client *redis.Client) repositories.CacheRepository {
	return &cacheRepositoryImpl{
		client: client,
	}
}

func (r *cacheRepositoryImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = r.client.Set(ctx, key, jsonValue, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

func (r *cacheRepositoryImpl) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get cache value: %w", err)
	}

	return val, nil
}

func (r *cacheRepositoryImpl) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache value: %w", err)
	}

	return nil
}

func (r *cacheRepositoryImpl) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return result > 0, nil
}

type MessageSentInfo struct {
	ExternalMessageID string    `json:"external_message_id"`
	SentAt            time.Time `json:"sent_at"`
}

func (r *cacheRepositoryImpl) SetMessageSent(ctx context.Context, messageID, externalMessageID string, sentAt time.Time) error {
	key := fmt.Sprintf("message_sent:%s", messageID)

	info := MessageSentInfo{
		ExternalMessageID: externalMessageID,
		SentAt:            sentAt,
	}

	return r.Set(ctx, key, info, 24*time.Hour)
}

func (r *cacheRepositoryImpl) GetMessageSent(ctx context.Context, messageID string) (string, time.Time, error) {
	key := fmt.Sprintf("message_sent:%s", messageID)

	val, err := r.Get(ctx, key)
	if err != nil {
		return "", time.Time{}, err
	}

	var info MessageSentInfo
	err = json.Unmarshal([]byte(val), &info)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to unmarshal message sent info: %w", err)
	}

	return info.ExternalMessageID, info.SentAt, nil
}

func (r *cacheRepositoryImpl) SetSchedulerStatus(ctx context.Context, status string) error {
	key := "scheduler_status"
	return r.Set(ctx, key, status, 0) // No expiration
}

func (r *cacheRepositoryImpl) GetSchedulerStatus(ctx context.Context) (string, error) {
	key := "scheduler_status"
	return r.Get(ctx, key)
}
