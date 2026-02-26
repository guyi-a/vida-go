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

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 201 {object} response.Response{data=dto.UserInfo} "注册成功"
// @Failure 400 {object} response.ErrorResponse "请求参数无效"
// @Router /auth/register [post]
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

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取 JWT Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=dto.TokenData} "登录成功"
// @Failure 400 {object} response.ErrorResponse "请求参数无效"
// @Failure 401 {object} response.ErrorResponse "用户名或密码错误"
// @Router /auth/login [post]
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

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出（当前仅返回成功）
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "登出成功"
// @Failure 401 {object} response.ErrorResponse "未授权"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 目前不做 token 黑名单，仅返回成功
	response.OK(c, "登出成功", nil)
}

// Me 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.UserInfo} "获取成功"
// @Failure 401 {object} response.ErrorResponse "未授权"
// @Router /auth/me [get]
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
