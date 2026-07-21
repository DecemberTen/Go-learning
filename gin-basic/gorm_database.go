package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openGormDatabase 使用已有的 SQL 数据库连接池初始化 GORM。
// 参数 sqlDB：已经建立连接并配置好的数据库连接池。
// 返回值：初始化成功时返回 *gorm.DB，失败时返回错误。
func openGormDatabase(sqlDB *sql.DB) (*gorm.DB, error) {

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			// 超过 200 毫秒视为慢查询。
			SlowThreshold: 200 * time.Millisecond,
			// 只记录警告、慢查询和错误
			LogLevel: logger.Warn,
			// 不把正常的“记录不存在”刷成错误日志。
			IgnoreRecordNotFoundError: true,
			// 日志中保留 ?，不输出实际参数，减少敏感信息泄露。
			ParameterizedQueries: true,
			Colorful:             false,
		},
	)

	db, err := gorm.Open(
		gormMysql.New(gormMysql.Config{
			Conn: sqlDB,
		}),
		&gorm.Config{
			Logger: gormLogger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gorm open database failed: %w", err)
	}
	return db, nil
}
