package models

import "time"

// LinkDataDTO Структура данных о ссылке для слоя service
type LinkDataDTO struct {
	Link	string
	FullURL	string
	ExpTime	int
}

// LinkData Структура данных о ссылке для слоя repositories
type LinkDataDB struct {
	Link	string
	FullURL	string
	ExpTime	time.Duration
	Perm	bool
	Custom	bool
}

// LinksAmount Структура данных о ссылках пользователя
type LinksAmount struct {
	All		int
	Perm	int
	Custom	int
}
