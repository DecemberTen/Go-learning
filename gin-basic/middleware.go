package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func tokenMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")

		parts := strings.Fields(authorization)
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "Invalid access token",
				},
			)
			return
		}

		tokenText := parts[1]

		userID, role, err := parseAccessToken(tokenText, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid access token",
			})
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Set("user_role", role)
		c.Next()
	}
}

// adminMiddleware 验证当前登录用户是否拥有管理员角色。
// 参数：db 为查询最新用户信息的 GORM 数据库对象。
// 返回值：返回用于限制管理员访问的 Gin 中间件。
func adminMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDText, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found",
			})
			c.Abort()
			return
		}
		userID, ok := userIDText.(int64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user ID",
			})
			c.Abort()
			return
		}

		user, err := findUserByIDWithGorm(c.Request.Context(), db, userID)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User no longer exists",
			})
			c.Abort()
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User is not admin",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
