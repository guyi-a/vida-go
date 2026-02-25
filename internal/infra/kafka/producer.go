package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vida-go/internal/config"
	"vida-go/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var producer *kafka.Writer

// TranscodeTask 转码任务消息体
type TranscodeTask struct {
	VideoID    int64  `json:"video_id"`
	ObjectName string `json:"object_name"`
	Bucket     string `json:"bucket"`
	FileFormat string `json:"file_format"`
	FileSize   int64  `json:"file_size"`
}

// TranscodeResult 转码结果消息体
type TranscodeResult struct {
	VideoID  int64  `json:"video_id"`
	Status   string `json:"status"`
	PlayURL  string `json:"play_url,omitempty"`
	CoverURL string `json:"cover_url,omitempty"`
	Duration int    `json:"duration,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Error    string `json:"error,omitempty"`
}

// InitProducer 初始化 Kafka 生产者
func InitProducer(cfg *config.KafkaConfig) error {
	producer = &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	logger.Info("Kafka producer initialized",
		zap.Strings("brokers", cfg.Brokers),
	)

	return nil
}

// SendTranscodeTask 发送转码任务到 Kafka
func SendTranscodeTask(ctx context.Context, topic string, task *TranscodeTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal transcode task: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("video-%d", task.VideoID)),
		Value: payload,
	}

	if err := producer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to send transcode task: %w", err)
	}

	logger.Info("Transcode task sent",
		zap.Int64("video_id", task.VideoID),
		zap.String("topic", topic),
		zap.String("object", task.ObjectName),
	)

	return nil
}

// SendRaw 发送原始消息到指定 topic
func SendRaw(ctx context.Context, topic, key string, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	}

	if err := producer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to send kafka message: %w", err)
	}
	return nil
}

// CloseProducer 关闭生产者
func CloseProducer() error {
	if producer == nil {
		return nil
	}
	logger.Info("Kafka producer closed")
	return producer.Close()
}
