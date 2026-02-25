package handler

import (
	"errors"
	"strconv"

	"vida-go/internal/api/dto"
	"vida-go/internal/api/middleware"
	"vida-go/internal/api/response"
	"vida-go/internal/service"
	"vida-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FavoriteHandler struct {
	favoriteService *service.FavoriteService
}

func NewFavoriteHandler(favoriteService *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

// Favorite POST /api/v1/favorites/:video_id
func (h *FavoriteHandler) Favorite(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	info, totalFav, err := h.favoriteService.Favorite(userID, videoID)
	if err != nil {
		handleFavoriteError(c, err)
		return
	}

	response.OK(c, "点赞成功", gin.H{
		"favorite_id":    info.ID,
		"user_id":        info.UserID,
		"video_id":       info.VideoID,
		"created_at":     info.CreatedAt,
		"total_favorites": totalFav,
	})
}

// Unfavorite DELETE /api/v1/favorites/:video_id
func (h *FavoriteHandler) Unfavorite(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	totalFav, err := h.favoriteService.Unfavorite(userID, videoID)
	if err != nil {
		handleFavoriteError(c, err)
		return
	}

	response.OK(c, "取消点赞成功", gin.H{
		"user_id":        userID,
		"video_id":       videoID,
		"total_favorites": totalFav,
	})
}

// GetStatus GET /api/v1/favorites/:video_id/status
func (h *FavoriteHandler) GetStatus(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	isFav, total, err := h.favoriteService.GetStatus(userID, videoID)
	if err != nil {
		handleFavoriteError(c, err)
		return
	}

	response.OK(c, "查询点赞状态成功", gin.H{
		"is_favorited":   isFav,
		"video_id":       videoID,
		"total_favorites": total,
	})
}

// ListMyFavorites GET /api/v1/favorites/my/list
func (h *FavoriteHandler) ListMyFavorites(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	page, pageSize := parsePagination(c)

	data, err := h.favoriteService.ListByUser(userID, page, pageSize)
	if err != nil {
		logger.Error("Get my favorites failed", zap.Error(err))
		response.InternalError(c, "获取我的点赞列表失败")
		return
	}

	response.OK(c, "获取我的点赞列表成功", data)
}

// ListVideoFavorites GET /api/v1/favorites/video/:video_id/list
func (h *FavoriteHandler) ListVideoFavorites(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	page, pageSize := parsePagination(c)

	data, err := h.favoriteService.ListByVideo(videoID, page, pageSize)
	if err != nil {
		handleFavoriteError(c, err)
		return
	}

	response.OK(c, "获取视频点赞列表成功", data)
}

// BatchStatus POST /api/v1/favorites/batch/status
func (h *FavoriteHandler) BatchStatus(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var req dto.BatchFavoriteStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	statusMap, err := h.favoriteService.BatchCheckStatus(userID, req.VideoIDs)
	if err != nil {
		logger.Error("Batch favorite status failed", zap.Error(err))
		response.InternalError(c, "批量查询点赞状态失败")
		return
	}

	response.OK(c, "批量查询点赞状态成功", gin.H{
		"favorites_status": statusMap,
	})
}

// GetMyFavoritedVideos GET /api/v1/favorites/my/videos
func (h *FavoriteHandler) GetMyFavoritedVideos(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	page, pageSize := parsePagination(c)

	videoIDs, total, err := h.favoriteService.GetFavoritedVideoIDs(userID, page, pageSize)
	if err != nil {
		logger.Error("Get my favorited videos failed", zap.Error(err))
		response.InternalError(c, "获取我点赞的视频ID列表失败")
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	response.OK(c, "获取我点赞的视频ID列表成功", gin.H{
		"video_ids":   videoIDs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

func handleFavoriteError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAlreadyFavorited):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrNotFavorited):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrVideoNotFound):
		response.NotFound(c, err.Error())
	default:
		logger.Error("Favorite operation failed", zap.Error(err))
		response.InternalError(c, "操作失败，请稍后重试")
	}
}
