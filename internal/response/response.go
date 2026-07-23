package response

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorCode string

const (
	ErrorCodeInvalidRequest      ErrorCode = "INVALID_REQUEST"
	ErrorCodeInvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
	ErrorCodeUserAlreadyExists   ErrorCode = "USER_ALREADY_EXISTS"
	ErrorCodeInvalidAccessToken  ErrorCode = "INVALID_ACCESS_TOKEN"
	ErrorCodeInvalidRefreshToken ErrorCode = "INVALID_REFRESH_TOKEN"
	ErrorCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrorCodeInternal            ErrorCode = "INTERNAL_ERROR"
)

// RespondError 返回统一格式的 HTTP 错误响应。
// 参数：context 为 Gin 请求上下文；status 为 HTTP 状态码；code 为业务错误码；message 为错误说明。
// 返回值：无，错误信息直接写入 HTTP 响应。
func RespondError(
	context *gin.Context,
	status int,
	code ErrorCode,
	message string,
) {
	context.JSON(status, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

// AbortError 中断 Gin 中间件链并返回统一错误响应。
// 参数：context 为 Gin 请求上下文；status 为 HTTP 状态码；code 为业务错误码；message 为错误说明。
// 返回值：无，中间件链被中断并写入 HTTP 错误响应。
func AbortError(
	context *gin.Context,
	status int,
	code ErrorCode,
	message string,
) {
	context.AbortWithStatusJSON(
		status,
		ErrorResponse{
			Code:    code,
			Message: message,
		},
	)
}
