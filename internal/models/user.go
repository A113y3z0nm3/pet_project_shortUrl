package models

// Subscribe определяет подписку пользователя (1 - Sub, 2 - Default)
type Subscribe int

// ChoiceRole Меняет формат подписки на цифровой
func (r *Subscribe) ChoiceSubscribe(role string) Subscribe {
	switch role {
	case "sub":
		return Sub
	case "default":
		return Default
	default:
		return Default
	}
}

// ChoiceString Меняет формат подписки на строчный
func (r *Subscribe) ChoiceString() string {
	switch *r {
	case Sub:
		return "sub"
	case Default:
		return "default"
	default:
		return "default"
	}
}

// Типы пользователей
const (
	Sub 	Subscribe = iota + 1
	Default
)

// User структура пользователя для базы данных
type UserDB struct {
	ID			string `json:"id"`
	Username	string `json:"username"`
	FirstName	string `json:"first_name"`
	LastName	string `json:"last_name"`
	Subscribe	string `json:"role"`
	Password	string `json:"-"`
}

// JWTUserInfo список информации, которая будет представлена о пользователе в JWT
type JWTUserInfo struct {
	Username 	string		`json:"username"`
	Subscribe	Subscribe	`json:"subscribe"`
	Err			error
}

// SignInUserDTO структура пользователя для слоя service
type SignInUserDTO struct {
	Username	string		`json:"username"`
	FirstName	string		`json:"first_name"`
	LastName	string		`json:"last_name"`
	Subscribe	Subscribe	`json:"role"`
	Password	string		`json:"-"`
}

// SignUpUserDTO структура пользователя для слоя service
type SignUpUserDTO struct {
	Username	string		`json:"username" bson:"username"`
	FirstName	string		`json:"first_name" bson:"first_name"`
	LastName	string		`json:"last_name" bson:"last_name"`
	Subscribe	Subscribe	`json:"role" bson:"role"`
	Password	string		`json:"-" bson:"password"`
}
