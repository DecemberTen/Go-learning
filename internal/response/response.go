package response

import (
	"example.com/go-learning/internal/apperror"
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code    apperror.Code `json:"code"`
	Message string        `json:"message"`
}

type ErrorCode = apperror.Code

const (
	ErrorCodeInvalidRequest      = apperror.CodeInvalidRequest
	ErrorCodeNotFound            = apperror.CodeNotFound
	ErrorCodeConflict            = apperror.CodeConflict
	ErrorCodeInvalidCredentials  = apperror.CodeInvalidCredentials
	ErrorCodeUserAlreadyExists   = apperror.CodeUserAlreadyExists
	ErrorCodeInvalidAccessToken  = apperror.CodeInvalidAccessToken
	ErrorCodeInvalidRefreshToken = apperror.CodeInvalidRefreshToken
	ErrorCodeForbidden           = apperror.CodeForbidden
	ErrorCodeInternal            = apperror.CodeInternal
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
