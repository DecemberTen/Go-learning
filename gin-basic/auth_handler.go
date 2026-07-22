package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	gormDB    *gorm.DB
	jwtSecret string
}

func newAuthHandler(gormDB *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		// store:  store,
		// db:     db,
		gormDB:    gormDB,
		jwtSecret: jwtSecret,
	}
}

func (handler *AuthHandler) RegisterRoutes(router *gin.Engine) {

	productGroup := router.Group("/auth")

	productGroup.POST("/register", handler.register)
	productGroup.POST("/login", handler.login)
	productGroup.POST("/logout", handler.logout)
	productGroup.POST("/refresh", handler.refresh)
	productGroup.GET("/profile", tokenMiddleware(handler.jwtSecret), handler.profile)
	productGroup.GET(
		"/admin-only",
		tokenMiddleware(handler.jwtSecret),
		adminMiddleware(handler.gormDB),
		handler.adminOnly,
	)

}

func (handler *AuthHandler) register(c *gin.Context) {
	// 注册逻辑

	var input RegisterRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := registerUser(c.Request.Context(), handler.gormDB, input)
	if errors.Is(err, ErrInvalidUsername) || errors.Is(err, ErrPasswordTooLong) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	if errors.Is(err, ErrUserAlreadyExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (handler *AuthHandler) login(c *gin.Context) {
	// 登录逻辑
	var input LoginRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := loginUser(c.Request.Context(), handler.gormDB, input, handler.jwtSecret)
	if errors.Is(err, ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (handler *AuthHandler) logout(c *gin.Context) {
	// 注销逻辑
	var input LogoutRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = logout(
		c.Request.Context(),
		handler.gormDB,
		input.RefreshToken,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

func (handler *AuthHandler) refresh(c *gin.Context) {
	// 刷新逻辑
	var input RefreshRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// accessToken, err := refreshAccessToken(
	response, err := rotateRefreshToken(
		c.Request.Context(),
		handler.gormDB,
		input.RefreshToken,
		handler.jwtSecret,
	)
	if errors.Is(err, ErrInvalidRefreshToken) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (handler *AuthHandler) profile(c *gin.Context) {
	// 获取用户信息
	idText, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "User ID not found",
		})
		return
	}
	id, ok := idText.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	user, err := findUserByIDWithGorm(c.Request.Context(), handler.gormDB, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User no longer exists",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}

// adminOnly 返回管理员专属内容。
// 参数：c 为当前 Gin 请求上下文。
// 返回值：无，结果直接写入 HTTP 响应。
func (handler *AuthHandler) adminOnly(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin access granted",
	})
}
