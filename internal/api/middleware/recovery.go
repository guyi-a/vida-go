package middleware

import (
	"net/http"

	"vida-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 恢复中间件，捕获panic
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic日志
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// 返回500错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})

				// 终止请求
				c.Abort()
			}
		}()

		c.Next()
	}
}
