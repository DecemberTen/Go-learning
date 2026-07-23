package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	appconfig "example.com/go-learning/internal/config"
	"github.com/go-sql-driver/mysql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 创建、配置并验证 MySQL 数据库连接池。
// 参数：ctx 控制连接验证的取消和超时；databaseConfig 提供数据库密码。
// 返回值：连接成功时返回 SQL 连接池；创建或验证失败时返回错误。
func Open(
	ctx context.Context,
	databaseConfig appconfig.DatabaseConfig,
) (*sql.DB, error) {
	mysqlConfig := mysql.Config{
		User:                 "root",
		Passwd:               databaseConfig.Password,
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "go_learning",
		ParseTime:            true,
		AllowNativePasswords: true,
		ClientFoundRows:      true,
	}

	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接池失败: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.PingContext(pingContext); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	return db, nil
}

// OpenGORM 使用已有 SQL 连接池初始化 GORM。
// 参数：sqlDB 为已建立并配置完成的数据库连接池。
// 返回值：初始化成功时返回 GORM 数据库对象；初始化失败时返回错误。
func OpenGORM(sqlDB *sql.DB) (*gorm.DB, error) {
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
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
		return nil, fmt.Errorf("初始化 GORM 失败: %w", err)
	}

	return db, nil
}
