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

// Favorite 点赞视频
// @Summary 点赞视频
// @Description 对指定视频点赞
// @Tags 点赞
// @Produce json
// @Security BearerAuth
// @Param video_id path int true "视频ID"
// @Success 200 {object} response.Response "点赞成功"
// @Failure 400 {object} response.ErrorResponse "已点赞"
// @Failure 404 {object} response.ErrorResponse "视频不存在"
// @Router /favorites/{video_id} [post]
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

// Unfavorite 取消点赞
// @Summary 取消点赞
// @Description 取消对指定视频的点赞
// @Tags 点赞
// @Produce json
// @Security BearerAuth
// @Param video_id path int true "视频ID"
// @Success 200 {object} response.Response "取消点赞成功"
// @Failure 400 {object} response.ErrorResponse "未点赞"
// @Router /favorites/{video_id} [delete]
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

// GetStatus 获取点赞状态
// @Summary 获取点赞状态
// @Description 查询是否点赞了指定视频
// @Tags 点赞
// @Produce json
// @Security BearerAuth
// @Param video_id path int true "视频ID"
// @Success 200 {object} response.Response "查询成功"
// @Router /favorites/{video_id}/status [get]
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

// ListMyFavorites 获取我的点赞列表
// @Summary 获取我的点赞列表
// @Description 获取当前用户的点赞记录列表
// @Tags 点赞
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /favorites/my/list [get]
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

// ListVideoFavorites 获取视频点赞列表
// @Summary 获取视频点赞列表
// @Description 获取指定视频的点赞用户列表
// @Tags 点赞
// @Produce json
// @Security BearerAuth
// @Param video_id path int true "视频ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /favorites/video/{video_id}/list [get]
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

// BatchStatus 批量查询点赞状态
// @Summary 批量查询点赞状态
// @Description 批量查询对多个视频的点赞状态
// @Tags 点赞
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BatchFavoriteStatusRequest true "视频ID列表"
// @Success 200 {object} response.Response "查询成功"
// @Router /favorites/batch/status [post]
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

// GetMyFavoritedVideos 获取我点赞的视频列表
// @Summary 获取我点赞的视频列表
// @Description 获取当前用户点赞过的视频详情列表
// @Tags 点赞
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.VideoListData} "获取成功"
// @Router /favorites/my/videos [get]
func (h *FavoriteHandler) GetMyFavoritedVideos(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	page, pageSize := parsePagination(c)

	data, err := h.favoriteService.GetFavoritedVideos(userID, page, pageSize)
	if err != nil {
		logger.Error("Get my favorited videos failed", zap.Error(err))
		response.InternalError(c, "获取点赞视频列表失败")
		return
	}

	response.OK(c, "获取点赞视频列表成功", data)
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
