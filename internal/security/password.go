package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	n      = 32768
	r      = 8
	p      = 1
	keyLen = 32
)

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash, err := scrypt.Key([]byte(password), salt, n, r, p, keyLen)
	if err != nil {
		return "", err
	}

	// return hex-encoded string with salt appended to password
	hashedPW := fmt.Sprintf("%s.%s", hex.EncodeToString(hash), hex.EncodeToString(salt))

	return hashedPW, nil
}

// ComparePasswords сравнивает два пароля (который ввел пользователь с тем, который лежит в базе данных)
func ComparePasswords(storedPassword string, suppliedPassword string) (bool, error) {
	encodeSalt := strings.Split(storedPassword, ".")

	salt, err := hex.DecodeString(encodeSalt[1])

	if err != nil {
		return false, fmt.Errorf("unable to verify user password")
	}

	hash, err := scrypt.Key([]byte(suppliedPassword), salt, n, r, p, keyLen)

	if err != nil {
		return false, fmt.Errorf("unable to verify user password")
	}

	return hex.EncodeToString(hash) == encodeSalt[0], nil
}
