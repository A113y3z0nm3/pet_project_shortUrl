package models

// CreateTokenDTO структура с информацией для генерации токена
type CreateTokenDTO struct {
	Username	string		`json:"username"`
	Subscribe	Subscribe	`json:"sub"`
}
