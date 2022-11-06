package services

import (
	"context"
	"crypto/rsa"
	"short_url/internal/models"
	"short_url/internal/security"
)

// TSConfig конфигурация для TokenService
type TSConfig struct {
	PrivateKey         *rsa.PrivateKey
	PublicKey          *rsa.PublicKey
	TokenExpirationSec int64
}

// TokenService отвечает за создание и валидацию access токена
type TokenService struct {
	privateKey         *rsa.PrivateKey
	publicKey          *rsa.PublicKey
	tokenExpirationSec int64
}

// NewTokenService фабрика для TokenService
func NewTokenService(c *TSConfig) *TokenService {
	return &TokenService{
		privateKey:         c.PrivateKey,
		publicKey:          c.PublicKey,
		tokenExpirationSec: c.TokenExpirationSec,
	}
}

// ValidateToken проверят, что токен валидный
func (s *TokenService) ValidateToken(ctx context.Context, token string) (models.JWTUserInfo, error) {
	var jwtUser models.JWTUserInfo

	claims, err := security.ValidateAccessToken(token, s.publicKey)

	if err != nil {
		return jwtUser, err
	}

	jwtUser.Username = claims.User.Username
	jwtUser.Subscribe = claims.User.Subscribe

	return jwtUser, nil
}

// CreateToken создает новый токен
func (s *TokenService) CreateToken(ctx context.Context, dto models.CreateTokenDTO) (string, error) {
	token, err := security.GenerateAccessToken(models.JWTUserInfo{Username: dto.Username, Subscribe: dto.Subscribe}, s.privateKey, s.tokenExpirationSec)

	if err != nil {
		return "", err
	}

	return token, nil
}
