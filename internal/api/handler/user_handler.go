package handler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"vida-go/internal/api/dto"
	"vida-go/internal/api/middleware"
	"vida-go/internal/api/response"
	"vida-go/internal/config"
	"vida-go/internal/infra/minio"
	"vida-go/internal/service"
	"vida-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
	authService *service.AuthService
}

func NewUserHandler(userService *service.UserService, authService *service.AuthService) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
	}
}

// GetProfile 获取用户公开主页信息（无需登录，供查看他人主页）
// @Summary 获取用户公开主页
// @Description 获取用户公开信息，用于展示他人主页
// @Tags 用户
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserFullInfo} "获取成功"
// @Failure 404 {object} response.ErrorResponse "用户不存在"
// @Router /users/{id}/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	info, err := h.userService.GetUserByID(targetID)
	if err != nil {
		handleUserError(c, err)
		return
	}

	response.OK(c, "获取成功", info)
}

// UploadAvatar 上传用户头像
// @Summary 上传用户头像
// @Description 上传头像图片，支持 jpg/png/gif/webp
// @Tags 用户
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "头像文件"
// @Success 200 {object} response.Response{data=dto.UserFullInfo} "上传成功"
// @Failure 400 {object} response.ErrorResponse "请求参数无效"
// @Router /users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "请先登录")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "请选择头像文件")
		return
	}

	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		response.BadRequest(c, "仅支持 jpg、png、gif、webp 格式")
		return
	}
	if file.Size > 2*1024*1024 {
		response.BadRequest(c, "头像大小不能超过 2MB")
		return
	}

	f, err := file.Open()
	if err != nil {
		response.InternalError(c, "打开文件失败")
		return
	}
	defer f.Close()

	objectName := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().Unix(), ext)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	} else if ext == ".gif" {
		contentType = "image/gif"
	} else if ext == ".webp" {
		contentType = "image/webp"
	}

	if _, err := minio.UploadFile(ctx, "user-avatars", objectName, f, file.Size, contentType); err != nil {
		logger.Error("Upload avatar failed", zap.Error(err))
		response.InternalError(c, "上传头像失败")
		return
	}

	minioCfg := config.GetMinIO()
	avatarURL := minio.GetPublicURL(minioCfg.Endpoint, minioCfg.UseSSL, "user-avatars", objectName)

	currentUser, _ := h.authService.GetCurrentUser(userID)
	req := dto.UserUpdateRequest{Avatar: &avatarURL}
	info, err := h.userService.UpdateUser(userID, currentUser, &req)
	if err != nil {
		handleUserError(c, err)
		return
	}

	response.OK(c, "头像上传成功", info)
}

// GetMe 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.UserInfo} "获取成功"
// @Failure 401 {object} response.ErrorResponse "未授权"
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "无法获取用户信息")
		return
	}

	info, err := h.userService.GetUserByID(userID)
	if err != nil {
		handleUserError(c, err)
		return
	}

	response.OK(c, "获取成功", info)
}

// GetUser 获取用户信息
// @Summary 获取指定用户信息
// @Description 根据用户ID获取用户信息（需要权限）
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserInfo} "获取成功"
// @Failure 403 {object} response.ErrorResponse "无权限"
// @Failure 404 {object} response.ErrorResponse "用户不存在"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	currentUser, _ := h.authService.GetCurrentUser(currentUserID)
	if currentUser == nil {
		response.Unauthorized(c, "无法获取用户信息")
		return
	}

	if currentUser.ID != targetID && currentUser.UserRole != "admin" {
		response.Forbidden(c, "没有权限查看该用户信息")
		return
	}

	info, err := h.userService.GetUserByID(targetID)
	if err != nil {
		handleUserError(c, err)
		return
	}

	response.OK(c, "获取成功", info)
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body dto.UserUpdateRequest true "更新信息"
// @Success 200 {object} response.Response{data=dto.UserInfo} "更新成功"
// @Failure 400 {object} response.ErrorResponse "请求参数无效"
// @Failure 403 {object} response.ErrorResponse "无权限"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	currentUser, _ := h.authService.GetCurrentUser(currentUserID)
	if currentUser == nil {
		response.Unauthorized(c, "无法获取用户信息")
		return
	}

	info, err := h.userService.UpdateUser(targetID, currentUser, &req)
	if err != nil {
		handleUserError(c, err)
		return
	}

	response.OK(c, "更新成功", info)
}

// DeleteUser 删除用户
// @Summary 删除用户（管理员）
// @Description 软删除指定用户
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 404 {object} response.ErrorResponse "用户不存在"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.userService.SoftDeleteUser(targetID); err != nil {
		handleUserError(c, err)
		return
	}

	response.OK(c, "删除成功", nil)
}

// RestoreUser 恢复用户
// @Summary 恢复用户（管理员）
// @Description 恢复已删除的用户
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "恢复成功"
// @Failure 404 {object} response.ErrorResponse "用户不存在"
// @Router /users/{id}/restore [post]
func (h *UserHandler) RestoreUser(c *gin.Context) {
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.userService.RestoreUser(targetID); err != nil {
		handleUserError(c, err)
		return
	}

	response.OK(c, "恢复成功", nil)
}

// SetAdmin 设置管理员
// @Summary 设置管理员（管理员）
// @Description 将指定用户设置为管理员
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserInfo} "设置成功"
// @Failure 404 {object} response.ErrorResponse "用户不存在"
// @Router /users/{id}/set-admin [post]
func (h *UserHandler) SetAdmin(c *gin.Context) {
	targetID, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	info, err := h.userService.SetAdminRole(targetID)
	if err != nil {
		handleUserError(c, err)
		return
	}

	response.OK(c, "设置管理员角色成功", info)
}

// ListUsers 获取用户列表
// @Summary 获取用户列表（管理员）
// @Description 分页获取用户列表
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param username query string false "用户名筛选"
// @Param user_role query string false "角色筛选"
// @Success 200 {object} response.Response "获取成功"
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var username, userRole *string
	if v := c.Query("username"); v != "" {
		username = &v
	}
	if v := c.Query("user_role"); v != "" {
		userRole = &v
	}

	data, err := h.userService.ListUsers(page, pageSize, username, userRole)
	if err != nil {
		logger.Error("List users failed", zap.Error(err))
		response.InternalError(c, "获取用户列表失败")
		return
	}

	response.OK(c, "获取成功", data)
}

// parseIDParam 从 URL 路径参数中解析 int64 ID
func parseIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func handleUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrUsernameExists):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrUserDeleted):
		response.Unauthorized(c, err.Error())
	default:
		if err.Error() == "没有权限修改该用户信息" {
			response.Forbidden(c, err.Error())
			return
		}
		logger.Error("User operation failed", zap.Error(err))
		response.InternalError(c, "操作失败，请稍后重试")
	}
}
