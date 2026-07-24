package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"example.com/go-learning/internal/apperror"
	"example.com/go-learning/internal/model"
	"example.com/go-learning/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const refreshTokenDuration = 7 * 24 * time.Hour

type AuthService struct {
	store     *repository.Store
	jwtSecret string
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type UserResult struct {
	ID        int64
	Username  string
	Email     string
	Role      string
	CreatedAt time.Time
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         UserResult
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// NewAuthService 创建认证业务 Service。
// 参数：store 为用户与 Refresh Token 数据访问 Store；jwtSecret 为 JWT 签名密钥。
// 返回值：返回可处理注册、登录、刷新和退出业务的 AuthService。
func NewAuthService(
	store *repository.Store,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

// hashPassword 使用 bcrypt 对原始密码进行单向哈希。
// 参数：password 为用户提交的原始密码。
// 返回值：返回 bcrypt 哈希字符串；生成失败时返回错误。
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", fmt.Errorf("哈希密码失败: %w", err)
	}

	return string(hash), nil
}

// comparePassword 验证原始密码是否匹配密码哈希。
// 参数：password 为原始密码；hash 为数据库中的 bcrypt 哈希。
// 返回值：匹配时返回 nil；不匹配或哈希无效时返回错误。
func comparePassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}

// Register 注册新用户。
// 参数：ctx 控制操作取消和超时；input 为注册数据。
// 返回值：注册成功返回公开用户信息；校验、哈希或创建失败时返回错误。
func (service *AuthService) Register(
	ctx context.Context,
	input RegisterInput,
) (UserResult, error) {
	username := strings.TrimSpace(input.Username)
	email := strings.ToLower(strings.TrimSpace(input.Email))

	usernameLength := utf8.RuneCountInString(username)
	if usernameLength < 3 || usernameLength > 50 {
		return UserResult{}, apperror.New(
			apperror.CodeInvalidRequest,
			"Invalid username or password",
		)
	}
	if len(input.Password) > 72 {
		return UserResult{}, apperror.New(
			apperror.CodeInvalidRequest,
			"Invalid username or password",
		)
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return UserResult{}, internalError(err)
	}

	user := model.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         "user",
	}
	err = service.store.CreateUser(ctx, &user)
	if errors.Is(err, repository.ErrUserAlreadyExists) {
		return UserResult{}, apperror.New(
			apperror.CodeUserAlreadyExists,
			"User already exists",
		)
	}
	if err != nil {
		return UserResult{}, internalError(err)
	}

	return userResult(user), nil
}

// Login 验证用户凭证并签发 Access Token 与 Refresh Token。
// 参数：ctx 控制操作取消和超时；input 为登录数据。
// 返回值：登录成功返回令牌和用户信息；凭证错误或内部操作失败时返回错误。
func (service *AuthService) Login(
	ctx context.Context,
	input LoginInput,
) (LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, err := service.store.FindUserByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return LoginResult{}, invalidCredentialsError()
	}
	if err != nil {
		return LoginResult{}, internalError(err)
	}

	err = comparePassword(input.Password, user.PasswordHash)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return LoginResult{}, invalidCredentialsError()
	}
	if err != nil {
		return LoginResult{}, internalError(err)
	}

	accessToken, err := generateAccessToken(
		user.ID,
		user.Role,
		service.jwtSecret,
	)
	if err != nil {
		return LoginResult{}, internalError(err)
	}

	refreshToken, err := service.issueRefreshToken(
		ctx,
		service.store,
		user.ID,
	)
	if err != nil {
		return LoginResult{}, internalError(err)
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenDuration.Seconds()),
		User:         userResult(*user),
	}, nil
}

// RefreshAccessToken 验证 Refresh Token 并签发新的 Access Token。
// 参数：ctx 控制查询取消和超时；tokenText 为客户端提交的 Refresh Token。
// 返回值：验证成功返回新的 Access Token；Token 无效或查询失败时返回错误。
func (service *AuthService) RefreshAccessToken(
	ctx context.Context,
	tokenText string,
) (string, error) {
	if strings.TrimSpace(tokenText) == "" {
		return "", invalidRefreshTokenError()
	}

	refreshToken, err := service.store.FindRefreshTokenByHash(
		ctx,
		hashRefreshToken(tokenText),
	)
	if errors.Is(err, repository.ErrNotFound) {
		return "", invalidRefreshTokenError()
	}
	if err != nil {
		return "", internalError(err)
	}
	if refreshToken.RevokedAt != nil ||
		!time.Now().Before(refreshToken.ExpiresAt) {
		return "", invalidRefreshTokenError()
	}

	user, err := service.store.FindUserByID(ctx, refreshToken.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return "", invalidRefreshTokenError()
	}
	if err != nil {
		return "", internalError(err)
	}

	accessToken, err := generateAccessToken(
		user.ID,
		user.Role,
		service.jwtSecret,
	)
	if err != nil {
		return "", internalError(err)
	}

	return accessToken, nil
}

// RotateRefreshToken 验证并轮换 Refresh Token，同时签发新的 Access Token。
// 参数：ctx 控制事务取消和超时；tokenText 为旧 Refresh Token。
// 返回值：成功返回新令牌；令牌无效或事务失败时返回错误。
func (service *AuthService) RotateRefreshToken(
	ctx context.Context,
	tokenText string,
) (RefreshResult, error) {
	if strings.TrimSpace(tokenText) == "" {
		return RefreshResult{}, invalidRefreshTokenError()
	}

	tokenHash := hashRefreshToken(tokenText)
	var response RefreshResult

	err := service.store.Transaction(
		ctx,
		func(transactionStore *repository.Store) error {
			currentToken, err :=
				transactionStore.FindRefreshTokenByHashForUpdate(
					ctx,
					tokenHash,
				)
			if errors.Is(err, repository.ErrNotFound) {
				return invalidRefreshTokenError()
			}
			if err != nil {
				return err
			}

			now := time.Now()
			if currentToken.RevokedAt != nil ||
				!now.Before(currentToken.ExpiresAt) {
				return invalidRefreshTokenError()
			}

			user, err := transactionStore.FindUserByID(
				ctx,
				currentToken.UserID,
			)
			if errors.Is(err, repository.ErrNotFound) {
				return invalidRefreshTokenError()
			}
			if err != nil {
				return err
			}

			accessToken, err := generateAccessToken(
				user.ID,
				user.Role,
				service.jwtSecret,
			)
			if err != nil {
				return err
			}

			err = transactionStore.RevokeRefreshToken(
				ctx,
				currentToken.ID,
				now,
			)
			if errors.Is(err, repository.ErrRefreshTokenNotActive) {
				return invalidRefreshTokenError()
			}
			if err != nil {
				return err
			}

			newRefreshToken, err := service.issueRefreshToken(
				ctx,
				transactionStore,
				user.ID,
			)
			if err != nil {
				return err
			}

			response = RefreshResult{
				AccessToken:  accessToken,
				RefreshToken: newRefreshToken,
				ExpiresIn:    int64(accessTokenDuration.Seconds()),
			}
			return nil
		},
	)
	if err != nil {
		if appError, exists := apperror.As(err); exists {
			return RefreshResult{}, appError
		}
		return RefreshResult{}, internalError(err)
	}

	return response, nil
}

// Logout 幂等撤销客户端提交的 Refresh Token。
// 参数：ctx 控制操作取消和超时；tokenText 为 Refresh Token 原文。
// 返回值：撤销或无需撤销时返回 nil；数据库操作失败时返回错误。
func (service *AuthService) Logout(
	ctx context.Context,
	tokenText string,
) error {
	if strings.TrimSpace(tokenText) == "" {
		return nil
	}

	err := service.store.RevokeRefreshTokenByHash(
		ctx,
		hashRefreshToken(tokenText),
		time.Now(),
	)
	if err != nil {
		return internalError(err)
	}

	return nil
}

// ParseAccessToken 解析并验证 Access Token。
// 参数：tokenText 为客户端提交的 JWT 字符串。
// 返回值：验证成功返回用户 ID 和角色；Token 无效时返回错误。
func (service *AuthService) ParseAccessToken(
	tokenText string,
) (int64, string, error) {
	userID, role, err := parseAccessToken(
		tokenText,
		service.jwtSecret,
	)
	if errors.Is(err, ErrJWTSecretMissing) {
		return 0, "", internalError(err)
	}
	if err != nil {
		return 0, "", apperror.New(
			apperror.CodeInvalidAccessToken,
			"Access token is invalid",
		)
	}

	return userID, role, nil
}

// FindUserByID 查询当前最新用户信息。
// 参数：ctx 控制查询取消和超时；id 为用户 ID。
// 返回值：查询成功返回用户实体；用户不存在或查询失败时返回错误。
func (service *AuthService) FindUserByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {
	user, err := service.store.FindUserByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.New(
			apperror.CodeInvalidAccessToken,
			"Current user no longer exists",
		)
	}
	if err != nil {
		return nil, internalError(err)
	}

	return user, nil
}

// invalidCredentialsError 创建统一登录凭证错误。
// 参数：无。
// 返回值：返回不暴露邮箱或密码哪一项错误的认证 AppError。
func invalidCredentialsError() error {
	return apperror.New(
		apperror.CodeInvalidCredentials,
		"Invalid email or password",
	)
}

// invalidRefreshTokenError 创建统一 Refresh Token 错误。
// 参数：无。
// 返回值：返回不暴露令牌失效具体原因的认证 AppError。
func invalidRefreshTokenError() error {
	return apperror.New(
		apperror.CodeInvalidRefreshToken,
		"Refresh token is invalid",
	)
}

// generateRefreshToken 生成安全随机 Refresh Token 及其 SHA-256 哈希。
// 参数：无。
// 返回值：返回 Token 原文和哈希；随机数生成失败时返回错误。
func generateRefreshToken() (string, string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("生成随机 Refresh Token 失败: %w", err)
	}

	tokenText := base64.RawURLEncoding.EncodeToString(randomBytes)
	return tokenText, hashRefreshToken(tokenText), nil
}

// hashRefreshToken 计算 Refresh Token 的 SHA-256 哈希。
// 参数：tokenText 为 Refresh Token 原文。
// 返回值：返回长度为 64 的十六进制哈希字符串。
func hashRefreshToken(tokenText string) string {
	hash := sha256.Sum256([]byte(tokenText))
	return hex.EncodeToString(hash[:])
}

// issueRefreshToken 生成 Refresh Token 并保存其哈希。
// 参数：ctx 控制操作取消和超时；store 为普通或事务 Store；userID 为用户 ID。
// 返回值：成功返回 Refresh Token 原文；生成或保存失败时返回错误。
func (service *AuthService) issueRefreshToken(
	ctx context.Context,
	store *repository.Store,
	userID int64,
) (string, error) {
	tokenText, tokenHash, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	refreshToken := model.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}
	if err := store.CreateRefreshToken(ctx, &refreshToken); err != nil {
		return "", err
	}

	return tokenText, nil
}

// userResult 将数据库用户实体转换为公开业务结果。
// 参数：user 为数据库用户实体。
// 返回值：返回不包含密码哈希的用户信息。
func userResult(user model.User) UserResult {
	return UserResult{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}
