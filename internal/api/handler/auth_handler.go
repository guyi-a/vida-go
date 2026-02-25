package handler

import (
	"errors"

	"vida-go/internal/api/dto"
	"vida-go/internal/api/middleware"
	"vida-go/internal/api/response"
	"vida-go/internal/service"
	"vida-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	userInfo, err := h.authService.Register(&req)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			response.BadRequest(c, err.Error())
			return
		}
		logger.Error("Register failed", zap.Error(err))
		response.InternalError(c, "注册失败，请稍后重试")
		return
	}

	response.Created(c, "注册成功", userInfo)
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	tokenData, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			response.Unauthorized(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrUserDeleted) {
			response.Unauthorized(c, err.Error())
			return
		}
		logger.Error("Login failed", zap.Error(err))
		response.InternalError(c, "登录失败，请稍后重试")
		return
	}

	response.OK(c, "登录成功", tokenData)
}

// Logout POST /api/v1/auth/logout（需要认证）
func (h *AuthHandler) Logout(c *gin.Context) {
	// 目前不做 token 黑名单，仅返回成功
	response.OK(c, "登出成功", nil)
}

// Me GET /api/v1/auth/me（需要认证）
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "无法获取用户信息")
		return
	}

	userInfo, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrUserDeleted) {
			response.Unauthorized(c, err.Error())
			return
		}
		logger.Error("Get current user failed", zap.Error(err), zap.Int64("user_id", userID))
		response.InternalError(c, "获取用户信息失败")
		return
	}

	response.OK(c, "获取成功", userInfo)
}
