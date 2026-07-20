package main

import (
	"database/sql"
	"fmt"

	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// openGormDatabase 使用已有的 SQL 数据库连接池初始化 GORM。
// 参数 sqlDB：已经建立连接并配置好的数据库连接池。
// 返回值：初始化成功时返回 *gorm.DB，失败时返回错误。
func openGormDatabase(sqlDB *sql.DB) (*gorm.DB, error) {

	db, err := gorm.Open(
		gormMysql.New(gormMysql.Config{
			Conn: sqlDB,
		}),
		&gorm.Config{},
	)
	if err != nil {
		return nil, fmt.Errorf("gorm open database failed: %w", err)
	}
	return db, nil
}
