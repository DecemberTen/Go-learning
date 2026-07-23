package repository

import (
	"context"

	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type Store struct {
	db *gorm.DB
}

// NewStore 创建使用指定 GORM 数据库对象的数据访问 Store。
// 参数：db 为普通数据库连接或事务对象。
// 返回值：返回可执行商品、用户和令牌数据操作的 Store。
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// Transaction 在同一个数据库事务中执行一组 Store 操作。
// 参数：ctx 控制事务取消和超时；operation 接收绑定到当前事务的 Store。
// 返回值：操作和提交成功返回 nil；回调失败或事务提交失败时返回错误。
func (store *Store) Transaction(
	ctx context.Context,
	operation func(transactionStore *Store) error,
) error {
	return store.db.WithContext(ctx).Transaction(
		func(transaction *gorm.DB) error {
			return operation(NewStore(transaction))
		},
	)
}
