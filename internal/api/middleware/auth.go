package middleware

import (
	"strings"

	"vida-go/internal/api/response"
	"vida-go/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID   = "currentUserID"
	ContextKeyUserRole = "currentUserRole"
)

// AuthRequired JWT 认证中间件，要求请求必须携带有效 Token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "无效或过期的认证令牌")
			c.Abort()
			return
		}

		// 将用户 ID 存入上下文，后续 Handler 可通过 c.GetInt64() 获取
		c.Set(ContextKeyUserID, claims.UserID)
		c.Next()
	}
}

// GetCurrentUserID 从 Gin Context 中获取当前登录用户 ID
func GetCurrentUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	userID, ok := val.(int64)
	return userID, ok
}

// UserRoleFetcher 用于获取用户角色的函数类型
type UserRoleFetcher func(userID int64) (string, error)

// AdminRequired 管理员权限中间件（必须在 AuthRequired 之后使用）
// roleFetcher 用于从数据库查询用户角色
func AdminRequired(roleFetcher UserRoleFetcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetCurrentUserID(c)
		if !ok {
			response.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		role, err := roleFetcher(userID)
		if err != nil {
			response.Unauthorized(c, "用户不存在")
			c.Abort()
			return
		}

		if role != "admin" {
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}

		c.Set(ContextKeyUserRole, role)
		c.Next()
	}
}

// extractToken 从 Authorization 头中提取 Bearer Token
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
