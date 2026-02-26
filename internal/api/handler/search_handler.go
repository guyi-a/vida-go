package handler

import (
	"strconv"

	"vida-go/internal/api/dto"
	"vida-go/internal/api/response"
	"vida-go/internal/service"
	"vida-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// SearchVideos 搜索视频
// @Summary 搜索视频
// @Description 根据关键词搜索视频，支持多种筛选条件
// @Tags 搜索
// @Produce json
// @Param q query string false "搜索关键词"
// @Param author_id query int false "作者ID"
// @Param video_id query int false "视频ID"
// @Param status query string false "视频状态"
// @Param sort query string false "排序方式: relevance, latest, hot" default(relevance)
// @Param start_time query int false "开始时间戳"
// @Param end_time query int false "结束时间戳"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.SearchVideoData} "搜索成功"
// @Failure 400 {object} response.ErrorResponse "请求参数无效"
// @Router /search/videos [get]
func (h *SearchHandler) SearchVideos(c *gin.Context) {
	var req dto.SearchVideoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	if v := c.Query("author_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.AuthorID = &id
		}
	}
	if v := c.Query("video_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.VideoID = &id
		}
	}
	if v := c.Query("start_time"); v != "" {
		if t, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.StartTime = &t
		}
	}
	if v := c.Query("end_time"); v != "" {
		if t, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.EndTime = &t
		}
	}

	data, err := h.searchService.SearchVideos(&req)
	if err != nil {
		logger.Error("Search videos failed", zap.Error(err))
		response.InternalError(c, "搜索失败")
		return
	}

	response.OK(c, "搜索成功", data)
}

// SyncVideosToES 同步视频到ES
// @Summary 同步视频到ES
// @Description 将数据库中的视频同步到 Elasticsearch
// @Tags 搜索
// @Produce json
// @Success 200 {object} response.Response "同步成功"
// @Failure 500 {object} response.ErrorResponse "同步失败"
// @Router /search/sync [post]
func (h *SearchHandler) SyncVideosToES(c *gin.Context) {
	success, failed, err := h.searchService.SyncVideosToES()
	if err != nil {
		logger.Error("Sync videos to ES failed", zap.Error(err))
		response.InternalError(c, "同步失败")
		return
	}

	response.OK(c, "同步完成", gin.H{
		"success": success,
		"failed":  failed,
	})
}
