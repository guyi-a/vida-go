package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"vida-go/internal/config"
	"vida-go/pkg/logger"

	"go.uber.org/zap"
)

// GetVideosIndexMapping 返回 videos 索引的 mapping（含 IK 中文分词）
func GetVideosIndexMapping() string {
	return `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"ik_max_word_analyzer": {
						"type": "custom",
						"tokenizer": "ik_max_word",
						"filter": ["lowercase"]
					},
					"ik_smart_analyzer": {
						"type": "custom",
						"tokenizer": "ik_smart",
						"filter": ["lowercase"]
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id": {"type": "long"},
				"author_id": {"type": "long"},
				"author_name": {"type": "keyword"},
				"title": {
					"type": "text",
					"analyzer": "ik_max_word",
					"search_analyzer": "ik_smart",
					"fields": {"keyword": {"type": "keyword", "ignore_above": 200}}
				},
				"description": {
					"type": "text",
					"analyzer": "ik_max_word",
					"search_analyzer": "ik_smart"
				},
				"status": {"type": "keyword"},
				"publish_time": {"type": "long"},
				"view_count": {"type": "long"},
				"favorite_count": {"type": "long"},
				"comment_count": {"type": "long"},
				"hot_score": {"type": "float"},
				"duration": {"type": "integer"},
				"created_at": {"type": "date", "format": "strict_date_optional_time||epoch_millis"},
				"updated_at": {"type": "date", "format": "strict_date_optional_time||epoch_millis"}
			}
		}
	}`
}

// EnsureVideosIndex 确保 videos 索引存在，不存在则创建
func EnsureVideosIndex(ctx context.Context) error {
	cfg := config.GetElasticsearch()
	indexName := cfg.Index["videos"]
	if indexName == "" {
		indexName = "videos"
	}

	exists, err := IndicesExists(ctx, indexName)
	if err != nil {
		return fmt.Errorf("check index exists: %w", err)
	}
	if exists {
		logger.Info("Elasticsearch videos index already exists", zap.String("index", indexName))
		return nil
	}

	body := bytes.NewReader([]byte(GetVideosIndexMapping()))
	resp, err := IndicesCreate(ctx, indexName, body)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("create index failed: %s", resp.String())
	}

	logger.Info("Elasticsearch videos index created", zap.String("index", indexName))
	return nil
}

// InitIndexes 初始化所有索引（启动时调用）
func InitIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return EnsureVideosIndex(ctx)
}
