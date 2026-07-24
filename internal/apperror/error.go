package apperror

import "errors"

type Code string

const (
	CodeInvalidRequest      Code = "INVALID_REQUEST"
	CodeNotFound            Code = "NOT_FOUND"
	CodeConflict            Code = "CONFLICT"
	CodeInvalidCredentials  Code = "INVALID_CREDENTIALS"
	CodeUserAlreadyExists   Code = "USER_ALREADY_EXISTS"
	CodeInvalidAccessToken  Code = "INVALID_ACCESS_TOKEN"
	CodeInvalidRefreshToken Code = "INVALID_REFRESH_TOKEN"
	CodeForbidden           Code = "FORBIDDEN"
	CodeInternal            Code = "INTERNAL_ERROR"
)

type Error struct {
	Code    Code
	Message string
	Cause   error
}

// Error 返回业务错误的公开文本。
// 参数：无。
// 返回值：返回可以安全展示的错误消息。
func (appError *Error) Error() string {
	return appError.Message
}

// Unwrap 返回被包装的内部错误。
// 参数：无。
// 返回值：返回 Cause，供 errors.Is 和 errors.As 继续检查错误链。
func (appError *Error) Unwrap() error {
	return appError.Cause
}

// New 创建不包含内部原因的业务错误。
// 参数：code 为稳定业务错误码；message 为允许返回给客户端的消息。
// 返回值：返回新的业务错误。
func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap 将内部错误包装成业务错误。
// 参数：code 为业务错误码；message 为公开消息；cause 为内部原始错误。
// 返回值：返回包含内部原因的业务错误。
func Wrap(code Code, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// As 从错误链中提取业务错误。
// 参数：err 为待检查的错误。
// 返回值：找到时返回业务错误和 true，否则返回 nil 和 false。
func As(err error) (*Error, bool) {
	var appError *Error
	if !errors.As(err, &appError) {
		return nil, false
	}

	return appError, true
}
