package logger

import (
	"context"
	"log/slog"
	"os"
)

type requestIDKey struct{}

// New 创建应用使用的 JSON 结构化 Logger。
// 参数：无。
// 返回值：返回默认日志级别为 Info 的 slog.Logger。
func New() *slog.Logger {
	handler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		},
	)

	return slog.New(handler)
}

// WithRequestID 将 Request ID 保存到标准库 Context。
// 参数：ctx 为当前请求 Context；requestID 为请求唯一标识。
// 返回值：返回包含 Request ID 的子 Context。
func WithRequestID(
	ctx context.Context,
	requestID string,
) context.Context {
	return context.WithValue(
		ctx,
		requestIDKey{},
		requestID,
	)
}

// FromContext 获取带有当前 Request ID 的 Logger。
// 参数：ctx 为可能包含 Request ID 的 Context。
// 返回值：存在 Request ID 时返回附带该字段的 Logger，否则返回默认 Logger。
func FromContext(ctx context.Context) *slog.Logger {
	requestID, exists := ctx.Value(
		requestIDKey{},
	).(string)

	if !exists || requestID == "" {
		return slog.Default()
	}

	return slog.Default().With(
		"request_id",
		requestID,
	)
}
