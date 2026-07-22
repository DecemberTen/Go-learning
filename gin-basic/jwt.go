package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenDuration = time.Hour

var ErrJWTSecretMissing = errors.New(
	"JWT签名密钥不能为空",
)

var ErrInvalidAccessToken = errors.New(
	"Access Token无效",
)

type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// generateAccessToken 为登录用户生成 Access Token。
// 参数：userID 为用户 ID，role 为用户角色，secret 为 JWT 签名密钥。
// 返回值：生成成功时返回 JWT 字符串；密钥缺失或签名失败时返回错误。
func generateAccessToken(
	userID int64,
	role string,
	secret string,
) (string, error) {
	if secret == "" {
		return "", ErrJWTSecretMissing
	}

	now := time.Now()

	claims := AccessTokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  strconv.FormatInt(userID, 10),
			Issuer:   "go-learning",
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(
				now.Add(accessTokenDuration),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenText, err := token.SignedString(
		[]byte(secret),
	)
	if err != nil {
		return "", fmt.Errorf(
			"生成Access Token失败: %w",
			err,
		)
	}

	return tokenText, nil
}

// parseAccessToken 解析并验证 Access Token。
// 参数：tokenText 为 JWT 字符串，secret 为 JWT 签名密钥。
// 返回值：验证成功时返回用户 ID 和角色；Token 无效时返回错误。
func parseAccessToken(
	tokenText string,
	secret string,
) (int64, string, error) {
	if secret == "" {
		return 0, "", ErrJWTSecretMissing
	}

	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenText,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer("go-learning"),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, "", fmt.Errorf(
			"%w: %v",
			ErrInvalidAccessToken,
			err,
		)
	}
	if !token.Valid {
		return 0, "", ErrInvalidAccessToken
	}

	userID, err := strconv.ParseInt(
		claims.Subject,
		10,
		64,
	)
	if err != nil || userID <= 0 {
		return 0, "", ErrInvalidAccessToken
	}

	if claims.Role == "" {
		return 0, "", ErrInvalidAccessToken
	}

	return userID, claims.Role, nil
}
