package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"vida-go/internal/config"
	"vida-go/internal/model"
	"vida-go/pkg/logger"

	"go.uber.org/zap"
)

// ESVideoDoc ES 视频文档结构
type ESVideoDoc struct {
	ID            int64   `json:"id"`
	AuthorID      int64   `json:"author_id"`
	AuthorName    string  `json:"author_name"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Status        string  `json:"status"`
	PublishTime   int64   `json:"publish_time"`
	ViewCount     int64   `json:"view_count"`
	FavoriteCount int64   `json:"favorite_count"`
	CommentCount  int64   `json:"comment_count"`
	HotScore      float64 `json:"hot_score"`
	Duration      int     `json:"duration"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func hotScore(view, fav, comment int64) float64 {
	return (float64(view)*0.5 + float64(fav)*2.0 + float64(comment)*1.5) / 1000
}

func videoToESDoc(v *model.Video, authorName string) *ESVideoDoc {
	pubTime := int64(0)
	if v.PublishTime != nil {
		pubTime = *v.PublishTime
	}
	return &ESVideoDoc{
		ID:            v.ID,
		AuthorID:      v.AuthorID,
		AuthorName:    authorName,
		Title:         v.Title,
		Description:   v.Description,
		Status:        v.Status,
		PublishTime:   pubTime,
		ViewCount:     v.ViewCount,
		FavoriteCount: v.FavoriteCount,
		CommentCount:  v.CommentCount,
		HotScore:      hotScore(v.ViewCount, v.FavoriteCount, v.CommentCount),
		Duration:      v.Duration,
		CreatedAt:     v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     v.UpdatedAt.Format(time.RFC3339),
	}
}

// SyncVideo 同步单个视频到 ES
func SyncVideo(ctx context.Context, v *model.Video, authorName string) error {
	cfg := config.GetElasticsearch()
	indexName := cfg.Index["videos"]
	if indexName == "" {
		indexName = "videos"
	}

	doc := videoToESDoc(v, authorName)
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	resp, err := Index(ctx, indexName, fmt.Sprintf("%d", v.ID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("index document failed: %s", resp.String())
	}

	logger.Debug("Video synced to ES", zap.Int64("video_id", v.ID))
	return nil
}

// DeleteVideo 从 ES 删除视频
func DeleteVideo(ctx context.Context, videoID int64) error {
	cfg := config.GetElasticsearch()
	indexName := cfg.Index["videos"]
	if indexName == "" {
		indexName = "videos"
	}

	resp, err := Delete(ctx, indexName, fmt.Sprintf("%d", videoID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() && resp.StatusCode != 404 {
		return fmt.Errorf("delete document failed: %s", resp.String())
	}
	return nil
}

// BulkSyncVideos 批量同步视频到 ES
func BulkSyncVideos(ctx context.Context, videos []model.Video, authorNames map[int64]string) (success, failed int, err error) {
	cfg := config.GetElasticsearch()
	indexName := cfg.Index["videos"]
	if indexName == "" {
		indexName = "videos"
	}

	var buf strings.Builder
	for _, v := range videos {
		authorName := authorNames[v.AuthorID]
		doc := videoToESDoc(&v, authorName)
		docBody, _ := json.Marshal(doc)

		buf.WriteString(fmt.Sprintf(`{"index":{"_index":"%s","_id":"%d"}}`, indexName, v.ID))
		buf.WriteString("\n")
		buf.Write(docBody)
		buf.WriteString("\n")
	}

	if buf.Len() == 0 {
		return 0, 0, nil
	}

	resp, err := Bulk(ctx, strings.NewReader(buf.String()))
	if err != nil {
		return 0, len(videos), err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return 0, len(videos), fmt.Errorf("bulk failed: %s", resp.String())
	}

	var bulkResp struct {
		Errors bool `json:"errors"`
		Items  []struct {
			Index struct {
				Status int `json:"status"`
			} `json:"index"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&bulkResp); err != nil {
		return len(videos), 0, nil
	}

	for _, item := range bulkResp.Items {
		if item.Index.Status >= 200 && item.Index.Status < 300 {
			success++
		} else {
			failed++
		}
	}

	logger.Info("Bulk sync to ES completed", zap.Int("success", success), zap.Int("failed", failed))
	return success, failed, nil
}
