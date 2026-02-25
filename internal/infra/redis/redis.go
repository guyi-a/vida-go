package redis

import (
	"context"
	"fmt"
	"time"

	"vida-go/internal/config"
	"vida-go/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var Client *redis.Client

// Init 初始化Redis客户端
func Init(cfg *config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info("Redis connected",
		zap.String("addr", cfg.Addr()),
		zap.Int("db", cfg.DB),
		zap.Int("pool_size", cfg.PoolSize),
	)

	return nil
}

// Close 关闭Redis连接
func Close() error {
	if Client == nil {
		return nil
	}
	logger.Info("Redis connection closed")
	return Client.Close()
}

// Get 获取Redis客户端实例
func Get() *redis.Client {
	return Client
}
