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

type RelationHandler struct {
	relationService *service.RelationService
}

func NewRelationHandler(relationService *service.RelationService) *RelationHandler {
	return &RelationHandler{relationService: relationService}
}

// Follow 关注用户
// @Summary 关注用户
// @Description 关注指定用户
// @Tags 关注
// @Produce json
// @Security BearerAuth
// @Param id path int true "被关注用户ID"
// @Success 200 {object} response.Response "关注成功"
// @Failure 400 {object} response.ErrorResponse "不能关注自己/已关注"
// @Router /relations/follow/{id} [post]
func (h *RelationHandler) Follow(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	result, err := h.relationService.Follow(currentUserID, targetID)
	if err != nil {
		handleRelationError(c, err)
		return
	}

	response.OK(c, "关注成功", result)
}

// Unfollow 取消关注
// @Summary 取消关注
// @Description 取消关注指定用户
// @Tags 关注
// @Produce json
// @Security BearerAuth
// @Param id path int true "被取消关注用户ID"
// @Success 200 {object} response.Response "取消关注成功"
// @Failure 400 {object} response.ErrorResponse "未关注该用户"
// @Router /relations/unfollow/{id} [post]
func (h *RelationHandler) Unfollow(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	result, err := h.relationService.Unfollow(currentUserID, targetID)
	if err != nil {
		handleRelationError(c, err)
		return
	}

	response.OK(c, "取消关注成功", result)
}

// GetFollowing 获取关注列表
// @Summary 获取用户关注列表
// @Description 获取指定用户的关注列表
// @Tags 关注
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /relations/following/{id} [get]
func (h *RelationHandler) GetFollowing(c *gin.Context) {
	userID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	page, pageSize := parsePagination(c)

	data, err := h.relationService.GetFollowingList(userID, page, pageSize)
	if err != nil {
		handleRelationError(c, err)
		return
	}

	response.OK(c, "获取关注列表成功", data)
}

// GetFollowers 获取粉丝列表
// @Summary 获取用户粉丝列表
// @Description 获取指定用户的粉丝列表
// @Tags 关注
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /relations/followers/{id} [get]
func (h *RelationHandler) GetFollowers(c *gin.Context) {
	userID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	page, pageSize := parsePagination(c)

	data, err := h.relationService.GetFollowerList(userID, page, pageSize)
	if err != nil {
		handleRelationError(c, err)
		return
	}

	response.OK(c, "获取粉丝列表成功", data)
}

// GetMyFollowing 获取我的关注列表
// @Summary 获取我的关注列表
// @Description 获取当前用户的关注列表
// @Tags 关注
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /relations/following/my/list [get]
func (h *RelationHandler) GetMyFollowing(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)
	page, pageSize := parsePagination(c)

	data, err := h.relationService.GetFollowingList(currentUserID, page, pageSize)
	if err != nil {
		handleRelationError(c, err)
		return
	}

	response.OK(c, "获取我的关注列表成功", data)
}

// GetMyFollowers 获取我的粉丝列表
// @Summary 获取我的粉丝列表
// @Description 获取当前用户的粉丝列表
// @Tags 关注
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /relations/followers/my/list [get]
func (h *RelationHandler) GetMyFollowers(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)
	page, pageSize := parsePagination(c)

	data, err := h.relationService.GetFollowerList(currentUserID, page, pageSize)
	if err != nil {
		handleRelationError(c, err)
		return
	}

	response.OK(c, "获取我的粉丝列表成功", data)
}

// GetFollowStatus 获取关注状态
// @Summary 获取关注状态
// @Description 查询是否关注了指定用户
// @Tags 关注
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "查询成功"
// @Router /relations/following/{id}/status [get]
func (h *RelationHandler) GetFollowStatus(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	isFollowing, err := h.relationService.GetFollowStatus(currentUserID, targetID)
	if err != nil {
		logger.Error("Get follow status failed", zap.Error(err))
		response.InternalError(c, "查询关注状态失败")
		return
	}

	response.OK(c, "查询关注状态成功", gin.H{
		"is_following": isFollowing,
		"follow_id":    targetID,
	})
}

// GetMutualFollows 获取互相关注列表
// @Summary 获取互相关注列表
// @Description 获取与当前用户互相关注的用户列表
// @Tags 关注
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response "获取成功"
// @Router /relations/mutual [get]
func (h *RelationHandler) GetMutualFollows(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)
	page, pageSize := parsePagination(c)

	data, err := h.relationService.GetMutualFollows(currentUserID, page, pageSize)
	if err != nil {
		logger.Error("Get mutual follows failed", zap.Error(err))
		response.InternalError(c, "获取互相关注列表失败")
		return
	}

	response.OK(c, "获取互相关注列表成功", data)
}

// BatchFollowStatus 批量查询关注状态
// @Summary 批量查询关注状态
// @Description 批量查询对多个用户的关注状态
// @Tags 关注
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BatchFollowStatusRequest true "用户ID列表"
// @Success 200 {object} response.Response "查询成功"
// @Router /relations/batch/status [post]
func (h *RelationHandler) BatchFollowStatus(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)

	var req dto.BatchFollowStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	statusMap, err := h.relationService.BatchCheckFollowStatus(currentUserID, req.UserIDs)
	if err != nil {
		logger.Error("Batch follow status failed", zap.Error(err))
		response.InternalError(c, "批量查询关注状态失败")
		return
	}

	response.OK(c, "批量查询关注状态成功", gin.H{
		"follow_status": statusMap,
	})
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func handleRelationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCannotFollowSelf):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrAlreadyFollowed):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrNotFollowed):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		response.NotFound(c, err.Error())
	default:
		logger.Error("Relation operation failed", zap.Error(err))
		response.InternalError(c, "操作失败，请稍后重试")
	}
}
