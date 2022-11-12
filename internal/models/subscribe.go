package models

import "time"

// Стоимость подписок
type SubPrice struct {
	// Недельная
	Weekly float64
	// Месячная
	Monthly float64
	// Годовая
	Yearly float64
}

// Структура с информацией о приобретенной пользователем подпиской
type SubInfo struct {
	Username	string
	Exp			time.Duration
}
