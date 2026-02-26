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

// Create 发表评论
// @Summary 发表评论
// @Description 对指定视频发表评论
// @Tags 评论
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path int true "视频ID"
// @Param request body dto.CommentCreateRequest true "评论内容"
// @Success 200 {object} response.Response "发表成功"
// @Failure 404 {object} response.ErrorResponse "视频不存在"
// @Router /comments/{video_id} [post]
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

// Update 更新评论
// @Summary 更新评论
// @Description 更新指定评论的内容
// @Tags 评论
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Param request body dto.CommentUpdateRequest true "更新内容"
// @Success 200 {object} response.Response "更新成功"
// @Failure 403 {object} response.ErrorResponse "无权限"
// @Failure 404 {object} response.ErrorResponse "评论不存在"
// @Router /comments/{id} [put]
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

// Delete 删除评论
// @Summary 删除评论
// @Description 删除指定评论
// @Tags 评论
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 403 {object} response.ErrorResponse "无权限"
// @Failure 404 {object} response.ErrorResponse "评论不存在"
// @Router /comments/{id} [delete]
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

// ListByVideo 获取视频评论列表
// @Summary 获取视频评论列表
// @Description 获取指定视频的评论列表
// @Tags 评论
// @Produce json
// @Security BearerAuth
// @Param video_id path int true "视频ID"
// @Param parent_id query int false "父评论ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /comments/video/{video_id} [get]
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

// ListReplies 获取评论回复列表
// @Summary 获取评论回复列表
// @Description 获取指定评论的回复列表
// @Tags 评论
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /comments/{id}/replies [get]
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

// ListMyComments 获取我的评论列表
// @Summary 获取我的评论列表
// @Description 获取当前用户发表的评论列表
// @Tags 评论
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /comments/my/list [get]
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
