package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type DatabaseConfig struct {
	Password string
}

type AuthConfig struct {
	JWTSecret string
}

type ServerConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

// Load 集中读取并校验应用启动所需配置。
// 参数：无，配置从当前进程的环境变量读取。
// 返回值：配置有效时返回 Config；必填配置缺失时返回错误。
func Load() (Config, error) {
	config := Config{
		Server: ServerConfig{
			Address:           os.Getenv("SERVER_ADDRESS"),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ShutdownTimeout:   10 * time.Second,
		},
		Database: DatabaseConfig{
			Password: os.Getenv("MYSQL_PASSWORD"),
		},
		Auth: AuthConfig{
			JWTSecret: os.Getenv("JWT_SECRET"),
		},
	}

	if config.Server.Address == "" {
		config.Server.Address = ":8080"
	}

	if config.Database.Password == "" {
		return Config{}, errors.New(
			"MYSQL_PASSWORD 未设置",
		)
	}

	if config.Auth.JWTSecret == "" {
		return Config{}, errors.New(
			"JWT_SECRET 未设置",
		)
	}

	return config, nil
}
