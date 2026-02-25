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

type VideoHandler struct {
	videoService *service.VideoService
}

func NewVideoHandler(videoService *service.VideoService) *VideoHandler {
	return &VideoHandler{videoService: videoService}
}

// Upload POST /api/v1/videos/upload
func (h *VideoHandler) Upload(c *gin.Context) {
	var req dto.VideoUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	file, err := c.FormFile("video_file")
	if err != nil {
		response.BadRequest(c, "请上传视频文件")
		return
	}

	allowedFormats := map[string]bool{
		".mp4": true, ".avi": true, ".mov": true,
		".mkv": true, ".flv": true, ".webm": true,
	}

	ext := ""
	if len(file.Filename) > 0 {
		for i := len(file.Filename) - 1; i >= 0; i-- {
			if file.Filename[i] == '.' {
				ext = file.Filename[i:]
				break
			}
		}
	}

	if !allowedFormats[ext] {
		response.BadRequest(c, "不支持的文件格式，支持: mp4, avi, mov, mkv, flv, webm")
		return
	}

	maxSize := int64(500 * 1024 * 1024) // 500MB
	if file.Size > maxSize || file.Size == 0 {
		response.BadRequest(c, "文件大小无效（不能为空，最大 500MB）")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)

	fileFormat := ext
	if len(fileFormat) > 0 && fileFormat[0] == '.' {
		fileFormat = fileFormat[1:]
	}

	f, err := file.Open()
	if err != nil {
		response.InternalError(c, "打开上传文件失败")
		return
	}
	defer f.Close()

	info, err := h.videoService.Upload(currentUserID, &req, f, file.Size, fileFormat)
	if err != nil {
		logger.Error("Upload video failed", zap.Error(err))
		response.InternalError(c, "上传视频失败: "+err.Error())
		return
	}

	response.OK(c, "视频上传成功，转码任务已提交", gin.H{
		"video_id": info.ID,
		"status":   info.Status,
	})
}

// GetFeed GET /api/v1/videos/feed（公开，不需要登录）
func (h *VideoHandler) GetFeed(c *gin.Context) {
	page, pageSize := parsePagination(c)

	data, err := h.videoService.GetFeed(page, pageSize)
	if err != nil {
		logger.Error("Get video feed failed", zap.Error(err))
		response.InternalError(c, "获取视频流失败")
		return
	}

	response.OK(c, "获取视频流成功", data)
}

// GetDetail GET /api/v1/videos/:id
func (h *VideoHandler) GetDetail(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	info, err := h.videoService.GetDetail(videoID)
	if err != nil {
		handleVideoError(c, err)
		return
	}

	response.OK(c, "获取视频详情成功", info)
}

// GetMyVideos GET /api/v1/videos/my/list
func (h *VideoHandler) GetMyVideos(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)
	page, pageSize := parsePagination(c)

	var status *string
	if v := c.Query("status"); v != "" {
		status = &v
	}

	data, err := h.videoService.GetMyVideos(currentUserID, page, pageSize, status)
	if err != nil {
		logger.Error("Get my videos failed", zap.Error(err))
		response.InternalError(c, "获取我的视频列表失败")
		return
	}

	response.OK(c, "获取我的视频列表成功", data)
}

// UpdateVideo PUT /api/v1/videos/:id
func (h *VideoHandler) UpdateVideo(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	var req dto.VideoUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)

	info, err := h.videoService.Update(videoID, currentUserID, &req)
	if err != nil {
		handleVideoError(c, err)
		return
	}

	response.OK(c, "更新视频成功", info)
}

// DeleteVideo DELETE /api/v1/videos/:id
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)

	if err := h.videoService.Delete(videoID, currentUserID); err != nil {
		handleVideoError(c, err)
		return
	}

	response.OK(c, "删除视频成功", nil)
}

func handleVideoError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrVideoNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrVideoNoPermission):
		response.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrNoFieldsToUpdate):
		response.BadRequest(c, err.Error())
	default:
		logger.Error("Video operation failed", zap.Error(err))
		response.InternalError(c, "操作失败，请稍后重试")
	}
}
