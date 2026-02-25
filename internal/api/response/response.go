package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一成功响应
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorInfo 错误详情
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// ErrorResponse 统一错误响应
type ErrorResponse struct {
	Error ErrorInfo `json:"error"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Fail(c *gin.Context, statusCode int, errType string, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorInfo{
			Code:    statusCode,
			Message: message,
			Type:    errType,
		},
	})
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, "BadRequest", message)
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, "Unauthorized", message)
}

func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, "Forbidden", message)
}

func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, "NotFound", message)
}

func InternalError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, "InternalServerError", message)
}
