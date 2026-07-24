package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	applogger "example.com/go-learning/internal/logger"
	"example.com/go-learning/internal/response"
	"github.com/gin-gonic/gin"
)

// generateRequestID 生成随机 Request ID。
// 参数：无。
// 返回值：成功时返回十六进制 Request ID；随机数生成失败时返回错误。
func generateRequestID() (string, error) {
	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

// NewRequestLogger 创建请求日志中间件。
// 参数：无。
// 返回值：返回生成 Request ID 并记录请求信息的 Gin 中间件。
func NewRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, err := generateRequestID()
		if err != nil {
			slog.Error(
				"生成 Request ID 失败",
				"error", err,
			)

			response.AbortError(
				c,
				http.StatusInternalServerError,
				response.ErrorCodeInternal,
				"Internal server error",
			)
			return
		}

		// 将 Request ID 返回给客户端。
		c.Header("X-Request-ID", requestID)

		// 基于当前 HTTP 请求 Context 创建带有 Request ID 的子 Context。
		requestContext := applogger.WithRequestID(
			c.Request.Context(),
			requestID,
		)

		// 把新的 Context 放回 HTTP Request。
		c.Request = c.Request.WithContext(
			requestContext,
		)

		logger := applogger.FromContext(requestContext)
		start := time.Now()

		logger.Info(
			"请求开始",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)

		// 执行后续中间件和 Handler。
		c.Next()

		logger.Info(
			"请求结束",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
