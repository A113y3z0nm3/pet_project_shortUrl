package security

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt"
	"short_url/internal/models"
	"time"
)

// AccessTokenCustomClaims кастомная структура для создания кастомного payload jwt
type AccessTokenCustomClaims struct {
	User models.JWTUserInfo `json:"user"`
	jwt.StandardClaims
}

// GenerateAccessToken генерирует токен доступа, который хранит имя и роль пользователя
func GenerateAccessToken(user models.JWTUserInfo, key *rsa.PrivateKey, exp int64) (string, error) {
	unixTime := time.Now().Unix()
	tokenExp := unixTime + exp

	claims := AccessTokenCustomClaims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  unixTime,
			ExpiresAt: tokenExp,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(key)

	if err != nil {
		return "", err
	}

	return ss, nil
}

// ValidateAccessToken возвращает claims пользователя с его именем и ролью, если токен валиден
func ValidateAccessToken(tokenString string, key *rsa.PublicKey) (*AccessTokenCustomClaims, error) {
	claims := &AccessTokenCustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("access token is invalid")
	}

	claims, ok := token.Claims.(*AccessTokenCustomClaims)

	if !ok {
		return nil, fmt.Errorf("access token valid but couldn't parse claims")
	}

	return claims, nil
}
