package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidUsername = errors.New(
	"用户名长度必须为3到50个字符",
)

var ErrPasswordTooLong = errors.New(
	"密码不能超过72字节",
)

var ErrInvalidCredentials = errors.New(
	"邮箱或密码错误",
)

// hashPassword 使用 bcrypt 对原始密码进行单向哈希。
// 参数：password 为用户提交的原始密码。
// 返回值：bcrypt 哈希字符串和可能出现的错误。
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password failed: %w", err)
	}
	return string(hash), nil
}

// registerUser 注册新用户。
// 参数：ctx 控制取消和超时，db 为 GORM 对象，input 为注册请求。
// 返回值：注册成功时返回公开用户信息；校验、哈希或创建失败时返回错误。
func registerUser(
	ctx context.Context,
	db *gorm.DB,
	input RegisterRequest,
) (UserResponse, error) {
	username := strings.TrimSpace(input.Username)
	email := strings.ToLower(strings.TrimSpace(input.Email))

	usernameLength := utf8.RuneCountInString(username)
	if usernameLength < 3 || usernameLength > 50 {
		return UserResponse{}, ErrInvalidUsername
	}

	if len(input.Password) > 72 {
		return UserResponse{}, ErrPasswordTooLong
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return UserResponse{}, fmt.Errorf("hash password failed: %w", err)
	}

	user := User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         "user",
	}

	err = createUserWithGorm(ctx, db, &user)

	if err != nil {
		return UserResponse{}, fmt.Errorf("create user failed: %w", err)
	}

	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

// comparePassword 使用 bcrypt 验证原始密码是否匹配密码哈希。
// 参数：password 为用户提交的原始密码，hash 为数据库保存的 bcrypt 哈希。
// 返回值：密码匹配返回 nil；不匹配或哈希异常时返回错误。
func comparePassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// loginUser 验证用户邮箱和密码。
// 参数：ctx 控制取消和超时，db 为 GORM 对象，input 为登录请求。
// 返回值：验证成功时返回登录响应；凭证错误或内部操作失败时返回错误。
func loginUser(
	ctx context.Context,
	db *gorm.DB,
	input LoginRequest,
	jwtSecret string,
) (LoginResponse, error) {

	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, err := findUserByEmailWithGorm(ctx, db, email)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return LoginResponse{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResponse{}, fmt.Errorf("query user failed: %w", err)
	}
	err = comparePassword(input.Password, user.PasswordHash)

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return LoginResponse{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResponse{}, fmt.Errorf("验证密码 failed: %w", err)
	}

	accessToken, err := generateAccessToken(user.ID, user.Role, jwtSecret)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("generate jwt failed: %w", err)
	}

	refreshToken, err := issueRefreshToken(ctx, db, user.ID)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("issue refresh token failed: %w", err)
	}

	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenDuration.Seconds()),
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// refreshAccessToken 验证 Refresh Token 并签发新的 Access Token。
// 参数：ctx 控制请求取消和超时，db 为 GORM 数据库对象，tokenText 为客户端提交的 Refresh Token，jwtSecret 为 JWT 密钥。
// 返回值：验证成功返回新的 Access Token；Token 无效或内部操作失败时返回错误。
func refreshAccessToken(
	ctx context.Context,
	db *gorm.DB,
	tokenText string,
	jwtSecret string,
) (string, error) {
	if strings.TrimSpace(tokenText) == "" {
		return "", ErrInvalidRefreshToken
	}

	tokenHash := hashRefreshToken(tokenText)

	refreshToken, err := findRefreshTokenByHashWithGorm(
		ctx,
		db,
		tokenHash,
	)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrInvalidRefreshToken
	}
	if err != nil {
		return "", err
	}

	if refreshToken.RevokedAt != nil {
		return "", ErrInvalidRefreshToken
	}

	if !time.Now().Before(refreshToken.ExpiresAt) {
		return "", ErrInvalidRefreshToken
	}

	user, err := findUserByIDWithGorm(
		ctx,
		db,
		refreshToken.UserID,
	)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrInvalidRefreshToken
	}
	if err != nil {
		return "", err
	}

	accessToken, err := generateAccessToken(
		user.ID,
		user.Role,
		jwtSecret,
	)
	if err != nil {
		return "", fmt.Errorf(
			"生成 Access Token 失败: %w",
			err,
		)
	}

	return accessToken, nil
}

// rotateRefreshToken 验证并轮换 Refresh Token，同时签发新的 Access Token。
// 参数：ctx 控制请求取消和超时，db 为 GORM 数据库对象，tokenText 为旧 Refresh Token，jwtSecret 为 JWT 密钥。
// 返回值：成功时返回新的 Access Token 和 Refresh Token；验证或数据库操作失败时返回错误。
func rotateRefreshToken(
	ctx context.Context,
	db *gorm.DB,
	tokenText string,
	jwtSecret string,
) (RefreshResponse, error) {
	if strings.TrimSpace(tokenText) == "" {
		return RefreshResponse{}, ErrInvalidRefreshToken
	}

	tokenHash := hashRefreshToken(tokenText)

	var response RefreshResponse

	err := db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			currentToken, err :=
				findRefreshTokenByHashForUpdateWithGorm(
					ctx,
					tx,
					tokenHash,
				)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidRefreshToken
			}
			if err != nil {
				return err
			}

			now := time.Now()

			if currentToken.RevokedAt != nil {
				return ErrInvalidRefreshToken
			}

			if !now.Before(currentToken.ExpiresAt) {
				return ErrInvalidRefreshToken
			}

			user, err := findUserByIDWithGorm(
				ctx,
				tx,
				currentToken.UserID,
			)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidRefreshToken
			}
			if err != nil {
				return err
			}

			accessToken, err := generateAccessToken(
				user.ID,
				user.Role,
				jwtSecret,
			)
			if err != nil {
				return fmt.Errorf(
					"生成 Access Token 失败: %w",
					err,
				)
			}

			err = revokeRefreshTokenWithGorm(
				ctx,
				tx,
				currentToken.ID,
				now,
			)
			if err != nil {
				return err
			}

			newRefreshToken, err := issueRefreshToken(
				ctx,
				tx,
				user.ID,
			)
			if err != nil {
				return err
			}

			response = RefreshResponse{
				AccessToken:  accessToken,
				RefreshToken: newRefreshToken,
				ExpiresIn: int64(
					accessTokenDuration.Seconds(),
				),
			}

			return nil
		},
	)
	if err != nil {
		return RefreshResponse{}, err
	}

	return response, nil
}

func logout(
	ctx context.Context,
	db *gorm.DB,
	tokenText string,
) error {
	if strings.TrimSpace(tokenText) == "" {
		return nil
	}
	tokenHash := hashRefreshToken(tokenText)

	err := revokeRefreshTokenByHashWithGorm(
		ctx,
		db,
		tokenHash,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}
