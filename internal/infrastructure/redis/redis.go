package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"message-sending-service/internal/infrastructure/config"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return rdb
}

func TestConnection(ctx context.Context, rdb *redis.Client) error {
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return nil
}
