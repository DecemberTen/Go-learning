package middleware

import (
	"errors"
	"net/http"
	"strings"

	"example.com/go-learning/internal/response"
	"example.com/go-learning/internal/service"
	"github.com/gin-gonic/gin"
)

// NewAuth 创建验证 Bearer Access Token 的认证中间件。
// 参数：authService 提供 Access Token 解析能力。
// 返回值：返回验证身份并把用户 ID、角色写入 Gin Context 的中间件。
func NewAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(context *gin.Context) {
		parts := strings.Fields(
			context.GetHeader("Authorization"),
		)
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			response.AbortError(
				context,
				http.StatusUnauthorized,
				response.ErrorCodeInvalidAccessToken,
				"Access token is invalid",
			)
			return
		}

		userID, role, err := authService.ParseAccessToken(parts[1])
		if err != nil {
			response.AbortError(
				context,
				http.StatusUnauthorized,
				response.ErrorCodeInvalidAccessToken,
				"Access token is invalid",
			)
			return
		}

		context.Set("user_id", userID)
		context.Set("user_role", role)
		context.Next()
	}
}

// NewAdmin 创建查询最新用户角色的管理员权限中间件。
// 参数：authService 提供当前用户查询能力。
// 返回值：返回只允许管理员继续访问的 Gin 中间件。
func NewAdmin(authService *service.AuthService) gin.HandlerFunc {
	return func(context *gin.Context) {
		userIDValue, exists := context.Get("user_id")
		userID, valid := userIDValue.(int64)
		if !exists || !valid {
			response.AbortError(
				context,
				http.StatusUnauthorized,
				response.ErrorCodeInvalidAccessToken,
				"Access token is invalid",
			)
			return
		}

		user, err := authService.FindUserByID(
			context.Request.Context(),
			userID,
		)
		if errors.Is(err, service.ErrNotFound) {
			response.AbortError(
				context,
				http.StatusUnauthorized,
				response.ErrorCodeInvalidAccessToken,
				"Current user no longer exists",
			)
			return
		}
		if err != nil {
			response.AbortError(
				context,
				http.StatusInternalServerError,
				response.ErrorCodeInternal,
				"Internal server error",
			)
			return
		}
		if user.Role != "admin" {
			response.AbortError(
				context,
				http.StatusForbidden,
				response.ErrorCodeForbidden,
				"Administrator permission is required",
			)
			return
		}

		context.Next()
	}
}
