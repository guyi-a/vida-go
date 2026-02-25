package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vida-go/internal/config"
	infraKafka "vida-go/internal/infra/kafka"
	infraMinio "vida-go/internal/infra/minio"
	"vida-go/internal/transcode"
	"vida-go/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	if err := logger.Init(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, cfg.Log.FilePath); err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer logger.Sync()

	if err := infraMinio.Init(&cfg.MinIO); err != nil {
		logger.Fatal("Failed to init minio", zap.Error(err))
	}

	if err := infraKafka.InitProducer(&cfg.Kafka); err != nil {
		logger.Fatal("Failed to init kafka producer", zap.Error(err))
	}
	defer infraKafka.CloseProducer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号，优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
		cancel()
	}()

	transcodeTopic := cfg.Kafka.Topics["video_transcode"]
	groupID := "vida-go-transcode-worker"

	logger.Info("Transcode worker started",
		zap.String("topic", transcodeTopic),
		zap.String("group", groupID),
		zap.Strings("brokers", cfg.Kafka.Brokers),
	)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          transcodeTopic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Info("Transcode worker stopped")
				return
			}
			logger.Error("Failed to read kafka message", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		var task infraKafka.TranscodeTask
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			logger.Error("Failed to unmarshal transcode task",
				zap.Error(err),
				zap.ByteString("value", msg.Value),
			)
			continue
		}

		logger.Info("Processing transcode task",
			zap.Int64("video_id", task.VideoID),
			zap.String("object", task.ObjectName),
		)

		if err := transcode.HandleTask(&task); err != nil {
			logger.Error("Transcode task failed",
				zap.Int64("video_id", task.VideoID),
				zap.Error(err),
			)
		} else {
			logger.Info("Transcode task completed",
				zap.Int64("video_id", task.VideoID),
			)
		}
	}
}
