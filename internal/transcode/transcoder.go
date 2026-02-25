package transcode

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"vida-go/internal/config"
	infraKafka "vida-go/internal/infra/kafka"
	infraMinio "vida-go/internal/infra/minio"
	"vida-go/pkg/logger"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

const (
	publicBucket = "public-videos"
	workDir      = "/tmp/vida-transcode"
)

// HandleTask 处理一个转码任务的完整流程：
//  1. 从 MinIO 下载原始视频
//  2. FFmpeg 转码为 mp4 (H.264 + AAC)
//  3. FFmpeg 截取封面图
//  4. 上传转码结果到 MinIO public-videos bucket
//  5. 发送转码结果消息到 Kafka
func HandleTask(task *infraKafka.TranscodeTask) error {
	taskDir := filepath.Join(workDir, fmt.Sprintf("%d", task.VideoID))
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return sendFailure(task.VideoID, fmt.Errorf("create work dir: %w", err))
	}
	defer os.RemoveAll(taskDir)

	srcFile := filepath.Join(taskDir, fmt.Sprintf("raw.%s", task.FileFormat))
	dstFile := filepath.Join(taskDir, "output.mp4")
	coverFile := filepath.Join(taskDir, "cover.jpg")

	logger.Info("Transcode task started",
		zap.Int64("video_id", task.VideoID),
		zap.String("object", task.ObjectName),
	)

	// 1. 从 MinIO 下载原始视频
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := downloadFromMinIO(ctx, task.Bucket, task.ObjectName, srcFile); err != nil {
		return sendFailure(task.VideoID, fmt.Errorf("download from minio: %w", err))
	}

	// 2. FFmpeg 转码
	if err := transcodeVideo(srcFile, dstFile); err != nil {
		return sendFailure(task.VideoID, fmt.Errorf("transcode: %w", err))
	}

	// 3. 截取封面
	if err := extractCover(dstFile, coverFile); err != nil {
		logger.Warn("Extract cover failed, skipping", zap.Error(err))
	}

	// 4. 探测视频信息
	probe, err := probeVideo(dstFile)
	if err != nil {
		logger.Warn("Probe video failed", zap.Error(err))
	}

	// 5. 上传转码后的视频和封面到 MinIO
	videoObjectName := fmt.Sprintf("videos/%d/video.mp4", task.VideoID)
	coverObjectName := fmt.Sprintf("videos/%d/cover.jpg", task.VideoID)

	if err := uploadToMinIO(ctx, publicBucket, videoObjectName, dstFile, "video/mp4"); err != nil {
		return sendFailure(task.VideoID, fmt.Errorf("upload video: %w", err))
	}

	var coverURL string
	if _, statErr := os.Stat(coverFile); statErr == nil {
		if err := uploadToMinIO(ctx, publicBucket, coverObjectName, coverFile, "image/jpeg"); err != nil {
			logger.Warn("Upload cover failed", zap.Error(err))
		} else {
			minioCfg := config.GetMinIO()
			coverURL = infraMinio.GetPublicURL(minioCfg.Endpoint, minioCfg.UseSSL, publicBucket, coverObjectName)
		}
	}

	minioCfg := config.GetMinIO()
	playURL := infraMinio.GetPublicURL(minioCfg.Endpoint, minioCfg.UseSSL, publicBucket, videoObjectName)

	// 6. 发送转码结果
	result := &infraKafka.TranscodeResult{
		VideoID:  task.VideoID,
		Status:   "published",
		PlayURL:  playURL,
		CoverURL: coverURL,
		Duration: probe.Duration,
		Width:    probe.Width,
		Height:   probe.Height,
	}

	return sendResult(result)
}

func downloadFromMinIO(ctx context.Context, bucket, objectName, destPath string) error {
	client := infraMinio.Get()
	obj, err := client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer obj.Close()

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.ReadFrom(obj); err != nil {
		return err
	}
	return nil
}

func transcodeVideo(srcFile, dstFile string) error {
	// H.264 + AAC, 分辨率保持不变, 中等质量 CRF 23
	args := []string{
		"-i", srcFile,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y",
		dstFile,
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg transcode failed: %w\noutput: %s", err, string(output))
	}

	logger.Info("FFmpeg transcode completed", zap.String("dst", dstFile))
	return nil
}

func extractCover(videoFile, coverFile string) error {
	// 截取第 1 秒的画面作为封面
	args := []string{
		"-i", videoFile,
		"-ss", "1",
		"-vframes", "1",
		"-q:v", "2",
		"-y",
		coverFile,
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg extract cover failed: %w\noutput: %s", err, string(output))
	}
	return nil
}

type videoProbe struct {
	Duration int
	Width    int
	Height   int
}

func probeVideo(videoFile string) (*videoProbe, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoFile,
	}

	cmd := exec.Command("ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return &videoProbe{}, fmt.Errorf("ffprobe failed: %w", err)
	}

	var data struct {
		Streams []struct {
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			Duration string `json:"duration"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	if err := json.Unmarshal(output, &data); err != nil {
		return &videoProbe{}, err
	}

	probe := &videoProbe{}

	if data.Format.Duration != "" {
		if dur, err := strconv.ParseFloat(data.Format.Duration, 64); err == nil {
			probe.Duration = int(dur)
		}
	}

	for _, s := range data.Streams {
		if s.Width > 0 && s.Height > 0 {
			probe.Width = s.Width
			probe.Height = s.Height
			break
		}
	}

	return probe, nil
}

func uploadToMinIO(ctx context.Context, bucket, objectName, filePath, contentType string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	_, err = infraMinio.UploadFile(ctx, bucket, objectName, f, info.Size(), contentType)
	return err
}

func sendResult(result *infraKafka.TranscodeResult) error {
	cfg := config.GetKafka()
	topic := cfg.Topics["video_uploaded"]

	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return infraKafka.SendRaw(ctx, topic, fmt.Sprintf("video-%d", result.VideoID), payload)
}

func sendFailure(videoID int64, originalErr error) error {
	logger.Error("Transcode task failed", zap.Int64("video_id", videoID), zap.Error(originalErr))

	result := &infraKafka.TranscodeResult{
		VideoID: videoID,
		Status:  "transcode_failed",
		Error:   originalErr.Error(),
	}

	if err := sendResult(result); err != nil {
		logger.Error("Failed to send failure result", zap.Error(err))
		return err
	}
	return originalErr
}

