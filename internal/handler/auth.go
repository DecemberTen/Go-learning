package handler

import (
	"net/http"

	"example.com/go-learning/internal/apperror"
	"example.com/go-learning/internal/response"
	"example.com/go-learning/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService  *service.AuthService
	authenticate gin.HandlerFunc
	requireAdmin gin.HandlerFunc
}

// NewAuthHandler 创建认证 HTTP Handler。
// 参数：authService 提供认证业务；authenticate 为登录校验中间件；requireAdmin 为管理员权限中间件。
// 返回值：返回可注册认证路由的 AuthHandler。
func NewAuthHandler(
	authService *service.AuthService,
	authenticate gin.HandlerFunc,
	requireAdmin gin.HandlerFunc,
) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		authenticate: authenticate,
		requireAdmin: requireAdmin,
	}
}

// RegisterRoutes 注册注册、登录、刷新、退出和用户资料路由。
// 参数：router 为 Gin 路由引擎。
// 返回值：无，路由直接注册到传入引擎。
func (handler *AuthHandler) RegisterRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	auth.POST("/register", handler.register)
	auth.POST("/login", handler.login)
	auth.POST("/logout", handler.logout)
	auth.POST("/refresh", handler.refresh)
	auth.GET(
		"/profile",
		handler.authenticate,
		handler.profile,
	)
	auth.GET(
		"/admin-only",
		handler.authenticate,
		handler.requireAdmin,
		handler.adminOnly,
	)
}

// register 创建新用户并返回公开用户信息。
// 参数：context 为当前 Gin 请求上下文，包含注册请求体。
// 返回值：无，注册结果直接写入 HTTP 响应。
func (handler *AuthHandler) register(context *gin.Context) {
	var input RegisterRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid request body",
		)
		return
	}

	user, err := handler.authService.Register(
		context.Request.Context(),
		service.RegisterInput{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
		},
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusCreated, userResponse(user))
}

// login 验证用户凭证并返回登录令牌。
// 参数：context 为当前 Gin 请求上下文，包含登录请求体。
// 返回值：无，登录结果直接写入 HTTP 响应。
func (handler *AuthHandler) login(context *gin.Context) {
	var input LoginRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid request body",
		)
		return
	}

	result, err := handler.authService.Login(
		context.Request.Context(),
		service.LoginInput{
			Email:    input.Email,
			Password: input.Password,
		},
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusOK, LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		User:         userResponse(result.User),
	})
}

// logout 撤销客户端提交的 Refresh Token。
// 参数：context 为当前 Gin 请求上下文，包含退出请求体。
// 返回值：无，退出成功返回 204，失败时写入错误响应。
func (handler *AuthHandler) logout(context *gin.Context) {
	var input LogoutRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid request body",
		)
		return
	}

	if err := handler.authService.Logout(
		context.Request.Context(),
		input.RefreshToken,
	); err != nil {
		response.HandleError(context, err)
		return
	}

	context.Status(http.StatusNoContent)
}

// refresh 验证并轮换 Refresh Token。
// 参数：context 为当前 Gin 请求上下文，包含刷新请求体。
// 返回值：无，刷新结果直接写入 HTTP 响应。
func (handler *AuthHandler) refresh(context *gin.Context) {
	var input RefreshRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid request body",
		)
		return
	}

	result, err := handler.authService.RotateRefreshToken(
		context.Request.Context(),
		input.RefreshToken,
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusOK, RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

// profile 查询并返回当前登录用户的最新资料。
// 参数：context 为当前 Gin 请求上下文，包含中间件写入的用户 ID。
// 返回值：无，用户资料或错误直接写入 HTTP 响应。
func (handler *AuthHandler) profile(context *gin.Context) {
	userIDValue, exists := context.Get("user_id")
	userID, valid := userIDValue.(int64)
	if !exists || !valid {
		response.HandleError(
			context,
			apperror.New(
				apperror.CodeInternal,
				"Internal server error",
			),
		)
		return
	}

	user, err := handler.authService.FindUserByID(
		context.Request.Context(),
		userID,
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}

// adminOnly 返回管理员专属内容。
// 参数：context 为当前 Gin 请求上下文。
// 返回值：无，结果直接写入 HTTP 响应。
func (handler *AuthHandler) adminOnly(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"message": "Admin access granted",
	})
}

// userResponse 将业务用户结果转换为 HTTP 用户响应。
// 参数：user 为 Service 返回的公开用户数据。
// 返回值：返回带 JSON 字段定义的用户响应。
func userResponse(user service.UserResult) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}
