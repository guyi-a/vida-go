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

// Follow POST /api/v1/relations/follow/:id
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

// Unfollow POST /api/v1/relations/unfollow/:id
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

// GetFollowing GET /api/v1/relations/following/:id
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

// GetFollowers GET /api/v1/relations/followers/:id
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

// GetMyFollowing GET /api/v1/relations/following/my/list
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

// GetMyFollowers GET /api/v1/relations/followers/my/list
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

// GetFollowStatus GET /api/v1/relations/following/:id/status
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

// GetMutualFollows GET /api/v1/relations/mutual
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

// BatchFollowStatus POST /api/v1/relations/batch/status
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
