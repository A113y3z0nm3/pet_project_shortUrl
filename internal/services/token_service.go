package services

import (
	"context"
	"crypto/rsa"
	"short_url/internal/models"
	"short_url/internal/security"
	log "short_url/pkg/logger"
)

// TSConfig конфигурация для TokenService
type TSConfig struct {
	PrivateKey			*rsa.PrivateKey
	PublicKey			*rsa.PublicKey
	TokenExpirationSec	int64
	Logger				*log.Log
}

// TokenService отвечает за создание и валидацию access токена
type TokenService struct {
	privateKey			*rsa.PrivateKey
	publicKey			*rsa.PublicKey
	tokenExpirationSec	int64
	logger				*log.Log
}

// NewTokenService фабрика для TokenService
func NewTokenService(c *TSConfig) *TokenService {
	return &TokenService{
		privateKey:         c.PrivateKey,
		publicKey:          c.PublicKey,
		tokenExpirationSec: c.TokenExpirationSec,
		logger:				c.Logger,
	}
}

// ValidateToken проверят, что токен валидный
func (s *TokenService) ValidateToken(ctx context.Context, token string) (models.JWTUserInfo, error) {
	ctx = log.ContextWithSpan(ctx, "ValidateToken")
	l := s.logger.WithContext(ctx)

	l.Debug("ValidateToken() started")
	defer l.Debug("ValidateToken() done")

	var jwtUser models.JWTUserInfo

	claims, err := security.ValidateAccessToken(token, s.publicKey)

	if err != nil {
		l.Errorf("Unable to validate access token. Error: %s", err)
		return jwtUser, err
	}

	jwtUser.Username = claims.User.Username
	jwtUser.Subscribe = claims.User.Subscribe

	return jwtUser, nil
}

// CreateToken создает новый токен
func (s *TokenService) CreateToken(ctx context.Context, dto models.CreateTokenDTO) (string, error) {
	ctx = log.ContextWithSpan(ctx, "CreateToken")
	l := s.logger.WithContext(ctx)

	l.Debug("CreateToken() started")
	defer l.Debug("CreateToken() done")

	token, err := security.GenerateAccessToken(models.JWTUserInfo{Username: dto.Username, Subscribe: dto.Subscribe}, s.privateKey, s.tokenExpirationSec)

	if err != nil {
		l.Errorf("Unable to create access token. Error: %s", err)
		return "", err
	}

	return token, nil
}
