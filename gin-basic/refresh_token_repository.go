package main

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// createRefreshTokenWithGorm 将 Refresh Token 记录保存到数据库。
// 参数：ctx 控制请求取消和超时，db 为 GORM 数据库对象，refreshToken 为待保存的令牌记录。
// 返回值：保存成功返回 nil，数据库操作失败时返回包含原因的错误。
func createRefreshTokenWithGorm(
	ctx context.Context,
	db *gorm.DB,
	refreshToken *RefreshToken,
) error {
	err := db.WithContext(ctx).
		Create(refreshToken).
		Error

	if err != nil {
		return fmt.Errorf("保存 Refresh Token 失败: %w", err)
	}

	return nil
}

// findRefreshTokenByHashWithGorm 根据 Token 哈希查询 Refresh Token 记录。
// 参数：ctx 控制请求取消和超时，db 为 GORM 数据库对象，tokenHash 为令牌原文计算出的哈希。
// 返回值：查询成功返回令牌记录；记录不存在或数据库查询失败时返回错误。
func findRefreshTokenByHashWithGorm(
	ctx context.Context,
	db *gorm.DB,
	tokenHash string,
) (*RefreshToken, error) {
	var refreshToken RefreshToken

	err := db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&refreshToken).
		Error

	if err != nil {
		return nil, fmt.Errorf(
			"查询 Refresh Token 失败: %w",
			err,
		)
	}

	return &refreshToken, nil
}

// findRefreshTokenByHashForUpdateWithGorm 查询并锁定 Refresh Token 记录。
// 参数：ctx 控制请求取消和超时，db 为当前事务对象，tokenHash 为令牌哈希。
// 返回值：查询成功返回被锁定的令牌记录；不存在或查询失败时返回错误。
func findRefreshTokenByHashForUpdateWithGorm(
	ctx context.Context,
	db *gorm.DB,
	tokenHash string,
) (*RefreshToken, error) {
	var refreshToken RefreshToken

	err := db.WithContext(ctx).
		Clauses(clause.Locking{
			Strength: "UPDATE",
		}).
		Where("token_hash = ?", tokenHash).
		First(&refreshToken).
		Error

	if err != nil {
		return nil, fmt.Errorf(
			"锁定 Refresh Token 失败: %w",
			err,
		)
	}

	return &refreshToken, nil
}

// revokeRefreshTokenWithGorm 将尚未撤销的 Refresh Token 标记为已撤销。
// 参数：ctx 控制请求取消和超时，db 为当前事务对象，tokenID 为待撤销记录 ID，revokedAt 为撤销时间。
// 返回值：撤销成功返回 nil；记录不存在、已经撤销或数据库操作失败时返回错误。
func revokeRefreshTokenWithGorm(
	ctx context.Context,
	db *gorm.DB,
	tokenID int64,
	revokedAt time.Time,
) error {
	result := db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where(
			"id = ? AND revoked_at IS NULL",
			tokenID,
		).
		Update("revoked_at", revokedAt)

	if result.Error != nil {
		return fmt.Errorf(
			"撤销 Refresh Token 失败: %w",
			result.Error,
		)
	}

	if result.RowsAffected != 1 {
		return ErrInvalidRefreshToken
	}

	return nil
}

// revokeRefreshTokenWithGorm 将尚未撤销的 Refresh Token 标记为已撤销。
// 参数：ctx 控制请求取消和超时，db 为当前事务对象，tokenID 为待撤销记录 ID，revokedAt 为撤销时间。
// 返回值：撤销成功返回 nil；记录不存在、已经撤销或数据库操作失败时返回错误。
func revokeRefreshTokenByHashWithGorm(
	ctx context.Context,
	db *gorm.DB,
	tokenHash string,
	revokedAt time.Time,
) error {
	result := db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where(
			"token_hash = ? AND revoked_at IS NULL",
			tokenHash,
		).
		Update("revoked_at", revokedAt)

	if result.Error != nil {
		return fmt.Errorf(
			"撤销 Refresh Token 失败: %w",
			result.Error,
		)
	}

	return nil
}
