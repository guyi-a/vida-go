package middleware

import (
	"time"

	"vida-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger Gin日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		duration := time.Since(start)

		// 记录日志
		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("duration", duration),
			zap.Int("body_size", c.Writer.Size()),
		)

		// 如果有错误，记录错误日志
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Request Error",
					zap.String("error", e.Error()),
					zap.Any("type", e.Type),
				)
			}
		}
	}
}
