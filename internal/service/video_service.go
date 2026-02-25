package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"vida-go/internal/api/dto"
	"vida-go/internal/config"
	infraKafka "vida-go/internal/infra/kafka"
	infraMinio "vida-go/internal/infra/minio"
	"vida-go/internal/model"
	"vida-go/internal/repository"
	"vida-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrVideoNotFound     = errors.New("视频不存在")
	ErrVideoNoPermission = errors.New("没有权限操作该视频")
	ErrNoFieldsToUpdate  = errors.New("没有需要更新的字段")
)

const rawVideoBucket = "raw-videos"

type VideoService struct {
	videoRepo *repository.VideoRepository
}

func NewVideoService(videoRepo *repository.VideoRepository) *VideoService {
	return &VideoService{videoRepo: videoRepo}
}

// Upload 上传视频：MinIO 存储 + Kafka 转码任务
func (s *VideoService) Upload(authorID int64, req *dto.VideoUploadRequest, fileReader io.Reader, fileSize int64, fileFormat string) (*dto.VideoInfo, error) {
	video := &model.Video{
		AuthorID:    authorID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		FileSize:    fileSize,
		FileFormat:  fileFormat,
	}

	if err := s.videoRepo.Create(video); err != nil {
		return nil, err
	}

	objectName := fmt.Sprintf("%d/%d.%s", authorID, video.ID, fileFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	contentType := "video/" + fileFormat
	if _, err := infraMinio.UploadFile(ctx, rawVideoBucket, objectName, fileReader, fileSize, contentType); err != nil {
		logger.Error("Upload to MinIO failed, rolling back video record",
			zap.Int64("video_id", video.ID), zap.Error(err))
		_ = s.videoRepo.SoftDelete(video.ID)
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	cfg := config.GetKafka()
	transcodeTopic := cfg.Topics["video_transcode"]

	task := &infraKafka.TranscodeTask{
		VideoID:    video.ID,
		ObjectName: objectName,
		Bucket:     rawVideoBucket,
		FileFormat: fileFormat,
		FileSize:   fileSize,
	}

	if err := infraKafka.SendTranscodeTask(ctx, transcodeTopic, task); err != nil {
		logger.Error("Send transcode task failed", zap.Int64("video_id", video.ID), zap.Error(err))
		_, _ = s.videoRepo.Update(video.ID, map[string]interface{}{"status": "upload_failed"})
		return nil, fmt.Errorf("提交转码任务失败: %w", err)
	}

	_, _ = s.videoRepo.Update(video.ID, map[string]interface{}{"status": "transcoding"})
	video.Status = "transcoding"

	return toVideoInfo(video, false), nil
}

// HandleTranscodeResult 处理 Kafka 消费者收到的转码结果
func (s *VideoService) HandleTranscodeResult(result *infraKafka.TranscodeResult) error {
	updates := map[string]interface{}{
		"status": result.Status,
	}

	if result.Status == "published" {
		updates["play_url"] = result.PlayURL
		updates["cover_url"] = result.CoverURL
		updates["duration"] = result.Duration
		updates["width"] = result.Width
		updates["height"] = result.Height
		now := time.Now().Unix()
		updates["publish_time"] = now
	}

	_, err := s.videoRepo.Update(result.VideoID, updates)
	if err != nil {
		return fmt.Errorf("update video %d after transcode failed: %w", result.VideoID, err)
	}

	logger.Info("Video transcode result processed",
		zap.Int64("video_id", result.VideoID),
		zap.String("status", result.Status),
	)

	return nil
}

// GetDetail 获取视频详情（自动增加观看次数）
func (s *VideoService) GetDetail(videoID int64) (*dto.VideoInfo, error) {
	video, err := s.videoRepo.GetByIDWithAuthor(videoID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	if video.Status == "published" {
		_ = s.videoRepo.IncrementViewCount(videoID)
		video.ViewCount++
	}

	return toVideoInfo(video, true), nil
}

// Update 更新视频信息（仅作者本人）
func (s *VideoService) Update(videoID, currentUserID int64, req *dto.VideoUpdateRequest) (*dto.VideoInfo, error) {
	if _, err := s.videoRepo.GetByIDAndAuthor(videoID, currentUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNoPermission
		}
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return nil, ErrNoFieldsToUpdate
	}

	video, err := s.videoRepo.Update(videoID, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	return toVideoInfo(video, false), nil
}

// Delete 软删除视频（仅作者本人）
func (s *VideoService) Delete(videoID, currentUserID int64) error {
	if _, err := s.videoRepo.GetByIDAndAuthor(videoID, currentUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVideoNoPermission
		}
		return err
	}

	if err := s.videoRepo.SoftDelete(videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVideoNotFound
		}
		return err
	}
	return nil
}

// GetFeed 获取视频流（已发布，含作者信息，不需要登录）
func (s *VideoService) GetFeed(page, pageSize int) (*dto.VideoListData, error) {
	skip := (page - 1) * pageSize
	status := "published"
	videos, total, err := s.videoRepo.ListVideos(skip, pageSize, nil, &status, nil, true)
	if err != nil {
		return nil, err
	}
	return buildVideoListData(videos, total, page, pageSize, true), nil
}

// GetMyVideos 获取当前用户的视频列表
func (s *VideoService) GetMyVideos(userID int64, page, pageSize int, status *string) (*dto.VideoListData, error) {
	skip := (page - 1) * pageSize
	videos, total, err := s.videoRepo.ListVideos(skip, pageSize, &userID, status, nil, false)
	if err != nil {
		return nil, err
	}
	return buildVideoListData(videos, total, page, pageSize, false), nil
}

// toVideoInfo 将 model.Video 转换为 dto.VideoInfo
func toVideoInfo(video *model.Video, includeAuthor bool) *dto.VideoInfo {
	info := &dto.VideoInfo{
		ID:            video.ID,
		AuthorID:      video.AuthorID,
		Title:         video.Title,
		Description:   video.Description,
		PlayURL:       video.PlayURL,
		CoverURL:      video.CoverURL,
		Duration:      video.Duration,
		FileSize:      video.FileSize,
		FileFormat:    video.FileFormat,
		Width:         video.Width,
		Height:        video.Height,
		Status:        video.Status,
		ViewCount:     video.ViewCount,
		FavoriteCount: video.FavoriteCount,
		CommentCount:  video.CommentCount,
		PublishTime:   video.PublishTime,
		CreatedAt:     video.CreatedAt,
		UpdatedAt:     video.UpdatedAt,
	}

	if includeAuthor && video.Author.ID != 0 {
		info.Author = &dto.AuthorBrief{
			ID:       video.Author.ID,
			Username: video.Author.UserName,
			Avatar:   video.Author.Avatar,
		}
	}

	return info
}

func buildVideoListData(videos []model.Video, total int64, page, pageSize int, includeAuthor bool) *dto.VideoListData {
	items := make([]dto.VideoInfo, 0, len(videos))
	for i := range videos {
		items = append(items, *toVideoInfo(&videos[i], includeAuthor))
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return &dto.VideoListData{
		Videos:     items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
