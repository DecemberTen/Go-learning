package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const refreshTokenDuration = 7 * 24 * time.Hour

var ErrInvalidRefreshToken = errors.New(
	"Refresh Token 无效",
)

// generateRefreshToken 生成安全的随机 Refresh Token 及其哈希。
// 参数：无。
// 返回值：返回 Token 原文、SHA-256 哈希；随机数生成失败时返回错误。
func generateRefreshToken() (string, string, error) {
	randomBytes := make([]byte, 32)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", "", fmt.Errorf("生成随机 Refresh Token 失败: %w", err)
	}

	tokenText := base64.RawURLEncoding.EncodeToString(randomBytes)
	tokenHash := hashRefreshToken(tokenText)

	return tokenText, tokenHash, nil
}

// hashRefreshToken 计算 Refresh Token 的 SHA-256 哈希。
// 参数：tokenText 为客户端提交的 Refresh Token 原文。
// 返回值：返回长度为 64 的十六进制哈希字符串。
func hashRefreshToken(tokenText string) string {
	hash := sha256.Sum256([]byte(tokenText))
	return hex.EncodeToString(hash[:])
}

// issueRefreshToken 为指定用户生成 Refresh Token 并保存其哈希。
// 参数：ctx 控制请求取消和超时，db 为 GORM 数据库对象，userID 为令牌所属用户。
// 返回值：成功时返回 Refresh Token 原文；生成或保存失败时返回错误。
func issueRefreshToken(
	ctx context.Context,
	db *gorm.DB,
	userID int64,
) (string, error) {
	tokenText, tokenHash, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	refreshToken := RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}

	err = createRefreshTokenWithGorm(
		ctx,
		db,
		&refreshToken,
	)
	if err != nil {
		return "", err
	}

	return tokenText, nil
}
