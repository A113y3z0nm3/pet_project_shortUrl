package config

import (
	"fmt"
	"log"
	"os"
	"short_url/internal/models"

	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt"
)

// LoadConfig загружает конфигурацию из env
func LoadConfig() (*models.Config, error) {
	// Инициализация конфигурации,
	// если был добавлен новый конфигу в models.Config, то
	// необходимо проинициализировать его тут, иначе будет nil pointer
	config := &models.Config{
		App:	&models.ConfigApp{},
		Price:	&models.ConfigPrice{},
		Log:	&models.ConfigLog{},
		HTTP:	&models.ConfigHTTP{},
		DB:		&models.ConfigDB{},
		RDB:	&models.ConfigRedis{},
		JWT:	&models.ConfigJWT{},
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("unable to load configuration. Error: %s", err)
	}

	privateFile, err := os.ReadFile("rsa_private.pem")

	if err != nil {
		log.Fatalf("could not read private key pem file: %e", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateFile)

	if err != nil {
		log.Fatalf("could not parse private key: %e", err)
	}

	publicFile, err := os.ReadFile("rsa_public.pem")

	if err != nil {
		log.Fatalf("could not read public key pem file: %e", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicFile)
	if err != nil {
		log.Fatalf("could not parse public key: %e", err)
	}

	config.JWT.PrivateKey = privateKey
	config.JWT.PublicKey = publicKey

	return config, nil
}
