package repository

import (
	"context"
	"errors"
	"fmt"

	"example.com/go-learning/internal/model"
	"github.com/go-sql-driver/mysql"
)

var ErrUserAlreadyExists = errors.New("用户名或邮箱已存在")

// CreateUser 将用户保存到数据库。
// 参数：ctx 控制操作取消和超时；user 为 Service 构造好的用户实体。
// 返回值：创建成功返回 nil；用户名或邮箱重复、数据库操作失败时返回错误。
func (store *Store) CreateUser(
	ctx context.Context,
	user *model.User,
) error {
	err := store.db.WithContext(ctx).Create(user).Error

	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		return ErrUserAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// FindUserByEmail 根据邮箱查询未被软删除的用户。
// 参数：ctx 控制查询取消和超时；email 为规范化后的邮箱。
// 返回值：查询成功返回用户；用户不存在或数据库查询失败时返回错误。
func (store *Store) FindUserByEmail(
	ctx context.Context,
	email string,
) (*model.User, error) {
	var user model.User

	err := store.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error
	if err != nil {
		return nil, fmt.Errorf("根据邮箱查询用户失败: %w", err)
	}

	return &user, nil
}

// FindUserByID 根据 ID 查询未被软删除的用户。
// 参数：ctx 控制查询取消和超时；id 为用户 ID。
// 返回值：查询成功返回用户；用户不存在或数据库查询失败时返回错误。
func (store *Store) FindUserByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {
	var user model.User

	err := store.db.WithContext(ctx).
		First(&user, id).
		Error
	if err != nil {
		return nil, fmt.Errorf("根据 ID 查询用户失败: %w", err)
	}

	return &user, nil
}
