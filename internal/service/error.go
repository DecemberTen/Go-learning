package service

import "example.com/go-learning/internal/apperror"

// internalError 将内部技术错误转换成不可泄露细节的业务错误。
// 参数：err 为 Repository、加密或令牌操作返回的内部错误。
// 返回值：返回公开消息固定为 Internal server error 的 AppError。
func internalError(err error) error {
	return apperror.Wrap(
		apperror.CodeInternal,
		"Internal server error",
		err,
	)
}
