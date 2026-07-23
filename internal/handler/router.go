package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFound 处理不存在的路由。
// 参数：context 为当前 Gin 请求上下文。
// 返回值：无，直接写入 404 JSON 响应。
func NotFound(context *gin.Context) {
	context.JSON(http.StatusNotFound, gin.H{
		"error": "Route not found",
	})
}

// MethodNotAllowed 处理路径存在但 HTTP 方法不支持的请求。
// 参数：context 为当前 Gin 请求上下文。
// 返回值：无，直接写入 405 JSON 响应。
func MethodNotAllowed(context *gin.Context) {
	context.JSON(http.StatusMethodNotAllowed, gin.H{
		"error": "Method not allowed",
	})
}
