package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"example.com/go-learning/internal/model"
	"gorm.io/gorm/clause"
)

var ErrRefreshTokenNotActive = errors.New("Refresh Token 不存在或已撤销")

// CreateRefreshToken 将 Refresh Token 记录保存到数据库。
// 参数：ctx 控制操作取消和超时；refreshToken 为待保存的令牌记录。
// 返回值：保存成功返回 nil；数据库操作失败时返回错误。
func (store *Store) CreateRefreshToken(
	ctx context.Context,
	refreshToken *model.RefreshToken,
) error {
	err := store.db.WithContext(ctx).Create(refreshToken).Error
	if err != nil {
		return fmt.Errorf("保存 Refresh Token 失败: %w", err)
	}

	return nil
}

// FindRefreshTokenByHash 根据哈希查询 Refresh Token。
// 参数：ctx 控制查询取消和超时；tokenHash 为令牌原文的哈希。
// 返回值：查询成功返回令牌记录；记录不存在或查询失败时返回错误。
func (store *Store) FindRefreshTokenByHash(
	ctx context.Context,
	tokenHash string,
) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken

	err := store.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&refreshToken).
		Error
	if err != nil {
		return nil, fmt.Errorf("查询 Refresh Token 失败: %w", err)
	}

	return &refreshToken, nil
}

// FindRefreshTokenByHashForUpdate 查询并锁定 Refresh Token 记录。
// 参数：ctx 控制查询取消和超时；tokenHash 为令牌哈希。
// 返回值：查询成功返回被锁定记录；记录不存在或查询失败时返回错误。
func (store *Store) FindRefreshTokenByHashForUpdate(
	ctx context.Context,
	tokenHash string,
) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken

	err := store.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", tokenHash).
		First(&refreshToken).
		Error
	if err != nil {
		return nil, fmt.Errorf("锁定 Refresh Token 失败: %w", err)
	}

	return &refreshToken, nil
}

// RevokeRefreshToken 根据 ID 撤销尚未撤销的 Refresh Token。
// 参数：ctx 控制操作取消和超时；tokenID 为令牌记录 ID；revokedAt 为撤销时间。
// 返回值：撤销成功返回 nil；记录不是活动状态或数据库操作失败时返回错误。
func (store *Store) RevokeRefreshToken(
	ctx context.Context,
	tokenID int64,
	revokedAt time.Time,
) error {
	result := store.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", tokenID).
		Update("revoked_at", revokedAt)

	if result.Error != nil {
		return fmt.Errorf("撤销 Refresh Token 失败: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrRefreshTokenNotActive
	}

	return nil
}

// RevokeRefreshTokenByHash 根据哈希幂等撤销 Refresh Token。
// 参数：ctx 控制操作取消和超时；tokenHash 为令牌哈希；revokedAt 为撤销时间。
// 返回值：数据库操作成功时返回 nil，即使令牌不存在或已经撤销；执行失败时返回错误。
func (store *Store) RevokeRefreshTokenByHash(
	ctx context.Context,
	tokenHash string,
	revokedAt time.Time,
) error {
	result := store.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", revokedAt)
	if result.Error != nil {
		return fmt.Errorf("撤销 Refresh Token 失败: %w", result.Error)
	}

	return nil
}
