package kafka

import (
	"context"
	"encoding/json"
	"time"

	"vida-go/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// ResultHandler 处理转码结果的回调函数
type ResultHandler func(result *TranscodeResult) error

// StartTranscodeResultConsumer 启动转码结果消费者（阻塞，需在 goroutine 中运行）
// ctx 取消后会自动停止
func StartTranscodeResultConsumer(ctx context.Context, brokers []string, topic, groupID string, handler ResultHandler) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("Failed to close kafka consumer", zap.Error(err))
		}
		logger.Info("Kafka transcode result consumer stopped")
	}()

	logger.Info("Kafka transcode result consumer started",
		zap.String("topic", topic),
		zap.String("group", groupID),
	)

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error("Failed to read kafka message", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		var result TranscodeResult
		if err := json.Unmarshal(msg.Value, &result); err != nil {
			logger.Error("Failed to unmarshal transcode result",
				zap.Error(err),
				zap.ByteString("value", msg.Value),
			)
			continue
		}

		logger.Info("Received transcode result",
			zap.Int64("video_id", result.VideoID),
			zap.String("status", result.Status),
		)

		if err := handler(&result); err != nil {
			logger.Error("Failed to handle transcode result",
				zap.Int64("video_id", result.VideoID),
				zap.Error(err),
			)
		}
	}
}
