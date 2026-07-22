package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrUserAlreadyExists = errors.New(
	"用户名或邮箱已存在",
)

// createUserWithGorm 将用户保存到数据库。
// 参数：ctx 控制取消和超时，db 为 GORM 对象，user 为 Service 构造好的用户。
// 返回值：创建成功返回 nil；用户重复或数据库操作失败时返回错误。
func createUserWithGorm(
	ctx context.Context,
	db *gorm.DB,
	user *User,
) error {
	err := db.WithContext(ctx).Create(user).Error

	var mysqlError *mysql.MySQLError

	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		return ErrUserAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("创建失败 %w", err)
	}
	return nil
}

// findUserByEmailWithGorm 根据邮箱查询未被软删除的用户。
// 参数：ctx 控制取消和超时，db 为 GORM 对象，email 为规范化后的邮箱。
// 返回值：查询成功返回用户；用户不存在或数据库查询失败时返回错误。

func findUserByEmailWithGorm(
	ctx context.Context,
	db *gorm.DB,
	email string,
) (*User, error) {
	var user User
	err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户失败 %w", err)
	}
	return &user, nil
}

func findUserByIDWithGorm(
	ctx context.Context,
	db *gorm.DB,
	id int64,
) (*User, error) {
	var user User
	err := db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户失败 %w", err)
	}
	return &user, nil
}
