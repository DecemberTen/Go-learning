package response

import (
	"net/http"

	"example.com/go-learning/internal/apperror"
	applogger "example.com/go-learning/internal/logger"
	"github.com/gin-gonic/gin"
)

// HandleError 将业务或内部错误转换成统一 HTTP 响应。
// 参数：context 为 Gin 请求上下文；err 为 Service 返回的错误。
// 返回值：无，错误日志和 HTTP 响应由该函数统一处理。
func HandleError(context *gin.Context, err error) {
	writeAppError(context, err, false)
}

// AbortAppError 将错误转换成统一 HTTP 响应并中断中间件链。
// 参数：context 为 Gin 请求上下文；err 为 Middleware 收到的错误。
// 返回值：无，中间件链被中断并写入错误响应。
func AbortAppError(context *gin.Context, err error) {
	writeAppError(context, err, true)
}

// writeAppError 解析错误、记录内部原因并写入统一响应。
// 参数：context 为 Gin 请求上下文；err 为待处理错误；abort 表示是否中断中间件链。
// 返回值：无，处理结果直接写入日志和 HTTP 响应。
func writeAppError(
	context *gin.Context,
	err error,
	abort bool,
) {
	appError, exists := apperror.As(err)
	if !exists {
		appError = apperror.Wrap(
			apperror.CodeInternal,
			"Internal server error",
			err,
		)
	}

	status := statusFromCode(appError.Code)
	if status == http.StatusInternalServerError {
		cause := appError.Cause
		if cause == nil {
			cause = err
		}

		applogger.FromContext(
			context.Request.Context(),
		).Error(
			"请求处理失败",
			"code", appError.Code,
			"error", cause,
		)
	}

	responseBody := ErrorResponse{
		Code:    appError.Code,
		Message: appError.Message,
	}

	if abort {
		context.AbortWithStatusJSON(status, responseBody)
		return
	}

	context.JSON(status, responseBody)
}

// statusFromCode 将业务错误码转换成 HTTP 状态码。
// 参数：code 为 Service 返回的业务错误码。
// 返回值：返回对应 HTTP 状态码；未知错误码返回 500。
func statusFromCode(code apperror.Code) int {
	switch code {
	case apperror.CodeInvalidRequest:
		return http.StatusBadRequest
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeConflict,
		apperror.CodeUserAlreadyExists:
		return http.StatusConflict
	case apperror.CodeInvalidCredentials,
		apperror.CodeInvalidAccessToken,
		apperror.CodeInvalidRefreshToken:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
