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

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// Create POST /api/v1/comments/:video_id
func (h *CommentHandler) Create(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	var req dto.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	info, err := h.commentService.Create(userID, videoID, &req)
	if err != nil {
		handleCommentError(c, err)
		return
	}

	response.OK(c, "发表评论成功", info)
}

// Update PUT /api/v1/comments/:id
func (h *CommentHandler) Update(c *gin.Context) {
	commentID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	var req dto.CommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	info, err := h.commentService.Update(commentID, userID, &req)
	if err != nil {
		handleCommentError(c, err)
		return
	}

	response.OK(c, "更新评论成功", info)
}

// Delete DELETE /api/v1/comments/:id
func (h *CommentHandler) Delete(c *gin.Context) {
	commentID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	_, err = h.commentService.Delete(commentID, userID)
	if err != nil {
		handleCommentError(c, err)
		return
	}

	response.OK(c, "删除评论成功", nil)
}

// ListByVideo GET /api/v1/comments/video/:video_id
func (h *CommentHandler) ListByVideo(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频ID")
		return
	}

	page, pageSize := parsePagination(c)

	var parentID *int64
	if v := c.Query("parent_id"); v != "" {
		pid, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			parentID = &pid
		}
	}

	data, err := h.commentService.ListByVideo(videoID, parentID, page, pageSize)
	if err != nil {
		handleCommentError(c, err)
		return
	}

	response.OK(c, "获取评论列表成功", data)
}

// ListReplies GET /api/v1/comments/:id/replies
func (h *CommentHandler) ListReplies(c *gin.Context) {
	commentID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	page, pageSize := parsePagination(c)

	data, err := h.commentService.ListReplies(commentID, page, pageSize)
	if err != nil {
		handleCommentError(c, err)
		return
	}

	response.OK(c, "获取回复列表成功", data)
}

// ListMyComments GET /api/v1/comments/my/list
func (h *CommentHandler) ListMyComments(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	page, pageSize := parsePagination(c)

	data, err := h.commentService.ListByUser(userID, page, pageSize)
	if err != nil {
		logger.Error("Get my comments failed", zap.Error(err))
		response.InternalError(c, "获取我的评论列表失败")
		return
	}

	response.OK(c, "获取我的评论列表成功", data)
}

func handleCommentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCommentNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrCommentNoPermission):
		response.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrVideoNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrParentNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrParentVideoMismatch):
		response.BadRequest(c, err.Error())
	default:
		logger.Error("Comment operation failed", zap.Error(err))
		response.InternalError(c, "操作失败，请稍后重试")
	}
}
