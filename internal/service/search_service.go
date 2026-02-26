package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"vida-go/internal/api/dto"
	"vida-go/internal/config"
	infraES "vida-go/internal/infra/elasticsearch"
	"vida-go/internal/model"
	"vida-go/internal/repository"
	"vida-go/pkg/logger"

	"go.uber.org/zap"
)

type SearchService struct {
	videoRepo *repository.VideoRepository
}

func NewSearchService(videoRepo *repository.VideoRepository) *SearchService {
	return &SearchService{videoRepo: videoRepo}
}

// SearchVideos 搜索视频（ES 优先，失败则降级到 DB）
func (s *SearchService) SearchVideos(req *dto.SearchVideoRequest) (*dto.SearchVideoData, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	data, err := s.searchFromES(req)
	if err != nil {
		logger.Warn("ES search failed, fallback to DB", zap.Error(err))
		return s.searchFromDB(req)
	}
	return data, nil
}

func (s *SearchService) searchFromES(req *dto.SearchVideoRequest) (*dto.SearchVideoData, error) {
	cfg := config.GetElasticsearch()
	indexName := cfg.Index["videos"]
	if indexName == "" {
		indexName = "videos"
	}

	query := s.buildESQuery(req)
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := infraES.Search(ctx, indexName, bytes.NewReader(queryJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("ES search error: %s", resp.String())
	}

	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source   struct {
					ID int64 `json:"id"`
				} `json:"_source"`
				Highlight map[string][]string `json:"highlight"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	videoIDs := make([]int64, 0, len(esResp.Hits.Hits))
	highlights := make(map[int64]map[string][]string)
	for _, h := range esResp.Hits.Hits {
		videoIDs = append(videoIDs, h.Source.ID)
		if len(h.Highlight) > 0 {
			highlights[h.Source.ID] = h.Highlight
		}
	}

	total := esResp.Hits.Total.Value
	if len(videoIDs) == 0 {
		return s.buildSearchData(nil, highlights, total, req.Page, req.PageSize), nil
	}

	videos, err := s.videoRepo.GetByIDsWithAuthor(videoIDs)
	if err != nil {
		return nil, err
	}

	videoMap := make(map[int64]*model.Video)
	for i := range videos {
		videoMap[videos[i].ID] = &videos[i]
	}

	ordered := make([]model.Video, 0, len(videoIDs))
	for _, id := range videoIDs {
		if v, ok := videoMap[id]; ok {
			ordered = append(ordered, *v)
		}
	}

	return s.buildSearchData(ordered, highlights, total, req.Page, req.PageSize), nil
}

func (s *SearchService) buildESQuery(req *dto.SearchVideoRequest) map[string]interface{} {
	boolQ := map[string]interface{}{
		"filter": []interface{}{
			map[string]interface{}{"term": map[string]interface{}{"status": "published"}},
		},
		"must": []interface{}{},
	}

	if strings.TrimSpace(req.Q) != "" {
		q := strings.TrimSpace(req.Q)
		if len(q) <= 2 {
			boolQ["should"] = []interface{}{
				map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  q,
						"fields": []string{"title^3", "description^1"},
						"type":   "best_fields",
						"operator": "or",
					},
				},
			}
			boolQ["minimum_should_match"] = 1
		} else {
			boolQ["must"] = append(boolQ["must"].([]interface{}),
				map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  q,
						"fields": []string{"title^3", "description^1"},
						"type":   "best_fields",
						"operator": "or",
						"minimum_should_match": "50%",
					},
				},
			)
		}
	}

	if req.AuthorID != nil {
		boolQ["filter"] = append(boolQ["filter"].([]interface{}),
			map[string]interface{}{"term": map[string]interface{}{"author_id": *req.AuthorID}})
	}
	if req.VideoID != nil {
		boolQ["filter"] = append(boolQ["filter"].([]interface{}),
			map[string]interface{}{"term": map[string]interface{}{"id": *req.VideoID}})
	}
	if req.StartTime != nil || req.EndTime != nil {
		rangeQ := map[string]interface{}{}
		if req.StartTime != nil {
			rangeQ["gte"] = *req.StartTime
		}
		if req.EndTime != nil {
			rangeQ["lte"] = *req.EndTime
		}
		boolQ["filter"] = append(boolQ["filter"].([]interface{}),
			map[string]interface{}{"range": map[string]interface{}{"publish_time": rangeQ}})
	}

	sortConfig := []interface{}{}
	switch req.Sort {
	case "time":
		sortConfig = append(sortConfig, map[string]interface{}{"publish_time": map[string]string{"order": "desc"}})
	case "hot":
		sortConfig = append(sortConfig, map[string]interface{}{"hot_score": map[string]string{"order": "desc"}})
	default:
		sortConfig = append(sortConfig, map[string]interface{}{"_score": map[string]string{"order": "desc"}})
		sortConfig = append(sortConfig, map[string]interface{}{"publish_time": map[string]string{"order": "desc"}})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQ,
		},
		"_source": []string{"id"},
		"from":    (req.Page - 1) * req.PageSize,
		"size":    req.PageSize,
		"sort":    sortConfig,
	}

	if strings.TrimSpace(req.Q) != "" {
		query["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"title":       map[string]interface{}{},
				"description": map[string]interface{}{},
			},
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
		}
	}

	return query
}

func (s *SearchService) buildSearchData(videos []model.Video, highlights map[int64]map[string][]string, total int64, page, pageSize int) *dto.SearchVideoData {
	items := make([]dto.SearchVideoInfo, 0, len(videos))
	for i := range videos {
		v := &videos[i]
		authorName := ""
		if v.Author.ID != 0 {
			authorName = v.Author.UserName
		}
		info := dto.SearchVideoInfo{
			ID:            v.ID,
			AuthorID:      v.AuthorID,
			AuthorName:    authorName,
			Title:         v.Title,
			Description:   v.Description,
			PlayURL:       v.PlayURL,
			CoverURL:      v.CoverURL,
			ViewCount:     v.ViewCount,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			PublishTime:   v.PublishTime,
			Highlight:     highlights[v.ID],
		}
		items = append(items, info)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	return &dto.SearchVideoData{
		Videos:     items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

func (s *SearchService) searchFromDB(req *dto.SearchVideoRequest) (*dto.SearchVideoData, error) {
	skip := (req.Page - 1) * req.PageSize
	status := "published"

	var authorID *int64
	if req.AuthorID != nil {
		authorID = req.AuthorID
	}
	var search *string
	if strings.TrimSpace(req.Q) != "" {
		q := strings.TrimSpace(req.Q)
		search = &q
	}

	videos, total, err := s.videoRepo.ListVideos(skip, req.PageSize, authorID, &status, search, true)
	if err != nil {
		return nil, err
	}

	if req.VideoID != nil {
		filtered := make([]model.Video, 0)
		for i := range videos {
			if videos[i].ID == *req.VideoID {
				filtered = append(filtered, videos[i])
				total = 1
				break
			}
		}
		videos = filtered
	}

	if req.Sort == "hot" {
		// 简单按热度排序（已在 DB 层可扩展）
		// 这里 ListVideos 默认按 created_at，如需 hot 需在 repo 加排序
	}

	return s.buildSearchData(videos, nil, total, req.Page, req.PageSize), nil
}

// SyncVideoToES 同步单个视频到 ES（转码完成后调用）
func (s *SearchService) SyncVideoToES(videoID int64) error {
	video, err := s.videoRepo.GetByIDWithAuthor(videoID)
	if err != nil {
		return err
	}
	if video.Status != "published" {
		return nil
	}

	authorName := ""
	if video.Author.ID != 0 {
		authorName = video.Author.UserName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return infraES.SyncVideo(ctx, video, authorName)
}

// SyncVideosToES 同步所有已发布视频到 ES
func (s *SearchService) SyncVideosToES() (success, failed int, err error) {
	status := "published"
	videos, _, err := s.videoRepo.ListVideos(0, 10000, nil, &status, nil, true)
	if err != nil {
		return 0, 0, err
	}

	if len(videos) == 0 {
		return 0, 0, nil
	}

	authorNames := make(map[int64]string)
	for i := range videos {
		if videos[i].Author.ID != 0 {
			authorNames[videos[i].AuthorID] = videos[i].Author.UserName
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return infraES.BulkSyncVideos(ctx, videos, authorNames)
}
