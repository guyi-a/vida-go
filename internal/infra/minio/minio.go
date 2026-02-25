package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"vida-go/internal/config"
	"vida-go/pkg/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

var client *minio.Client

// Init 初始化 MinIO 客户端并确保所有 Bucket 存在
func Init(cfg *config.MinIOConfig) error {
	var err error
	client, err = minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, bucket := range cfg.Buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("failed to check bucket %s: %w", bucket, err)
		}
		if !exists {
			if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
			logger.Info("MinIO bucket created", zap.String("bucket", bucket))
		}
	}

	// public-videos 需要公开读，供前端直接播放视频
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::public-videos/*"]}]}`
	if err := client.SetBucketPolicy(ctx, "public-videos", policy); err != nil {
		return fmt.Errorf("failed to set public policy for public-videos: %w", err)
	}
	logger.Info("MinIO public-videos bucket set to public-read")

	logger.Info("MinIO connected",
		zap.String("endpoint", cfg.Endpoint),
		zap.Int("buckets", len(cfg.Buckets)),
	)

	return nil
}

// Get 获取 MinIO 客户端实例
func Get() *minio.Client {
	return client
}

// UploadFile 上传文件到指定 Bucket
// 返回对象名（objectName）
func UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader, fileSize int64, contentType string) (string, error) {
	_, err := client.PutObject(ctx, bucket, objectName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}
	return objectName, nil
}

// GetPresignedURL 生成预签名下载 URL（有效期可配置）
func GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := client.PresignedGetObject(ctx, bucket, objectName, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}
	return presignedURL.String(), nil
}

// GetPublicURL 生成公开访问 URL（需要 Bucket 设置为 public-read）
func GetPublicURL(endpoint string, useSSL bool, bucket, objectName string) string {
	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint, bucket, objectName)
}
