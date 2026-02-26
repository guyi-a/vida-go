package elasticsearch

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"vida-go/internal/config"
	"vida-go/pkg/logger"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"
)

var client *elasticsearch.Client

// Init 初始化 Elasticsearch 客户端
func Init(cfg *config.ElasticsearchConfig) error {
	hosts := make([]string, 0, len(cfg.Hosts))
	for _, h := range cfg.Hosts {
		h = strings.TrimSpace(h)
		if h != "" && !strings.HasPrefix(h, "http") {
			h = "http://" + h
		}
		hosts = append(hosts, h)
	}

	if len(hosts) == 0 {
		return fmt.Errorf("elasticsearch hosts is empty")
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     hosts,
		RetryOnStatus: []int{502, 503, 504},
		MaxRetries:    3,
		RetryBackoff:  func(i int) time.Duration { return time.Duration(i) * time.Second },
	})
	if err != nil {
		return fmt.Errorf("create elasticsearch client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := es.Ping(es.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("elasticsearch ping failed: %s", resp.String())
	}

	client = es
	logger.Info("Elasticsearch connected", zap.Strings("hosts", hosts))
	return nil
}

// Get 获取 ES 客户端
func Get() *elasticsearch.Client {
	return client
}

// Search 执行搜索（body 为 JSON 字符串）
func Search(ctx context.Context, index string, body io.Reader) (*esapi.Response, error) {
	if client == nil {
		return nil, fmt.Errorf("elasticsearch client not initialized")
	}
	return client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex(index),
		client.Search.WithBody(body),
	)
}

// Index 索引文档
func Index(ctx context.Context, index, id string, body io.Reader) (*esapi.Response, error) {
	if client == nil {
		return nil, fmt.Errorf("elasticsearch client not initialized")
	}
	return client.Index(
		index,
		body,
		client.Index.WithContext(ctx),
		client.Index.WithDocumentID(id),
	)
}

// Delete 删除文档
func Delete(ctx context.Context, index, id string) (*esapi.Response, error) {
	if client == nil {
		return nil, fmt.Errorf("elasticsearch client not initialized")
	}
	return client.Delete(
		index,
		id,
		client.Delete.WithContext(ctx),
	)
}

// IndicesCreate 创建索引
func IndicesCreate(ctx context.Context, index string, body io.Reader) (*esapi.Response, error) {
	if client == nil {
		return nil, fmt.Errorf("elasticsearch client not initialized")
	}
	return client.Indices.Create(
		index,
		client.Indices.Create.WithContext(ctx),
		client.Indices.Create.WithBody(body),
	)
}

// IndicesExists 检查索引是否存在
func IndicesExists(ctx context.Context, index string) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("elasticsearch client not initialized")
	}
	resp, err := client.Indices.Exists(
		[]string{index},
		client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, err
	}
	return !resp.IsError() && resp.StatusCode == 200, nil
}

// Bulk 批量操作
func Bulk(ctx context.Context, body io.Reader) (*esapi.Response, error) {
	if client == nil {
		return nil, fmt.Errorf("elasticsearch client not initialized")
	}
	return client.Bulk(
		body,
		client.Bulk.WithContext(ctx),
	)
}

// Close 关闭连接
func Close() error {
	client = nil
	logger.Info("Elasticsearch client closed")
	return nil
}
