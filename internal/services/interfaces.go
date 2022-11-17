package services

import (
	"time"
	"context"
	"short_url/internal/models"
)

// authRepository Интерфейс к репозиторию хранения данных пользователей
type authRepository interface {
	FindByUsername(ctx context.Context, username string) (models.UserDB, error)
	CreateUser(ctx context.Context, user models.UserDB) error
}

// linkRepository Интерфейс к репозиторию управления ссылками
type linkRepository interface {
	CreateLink(ctx context.Context, link, username, fullUrl string, exp time.Duration, custom bool) (models.LinkDataDB, error)
	DeleteLink(ctx context.Context, link, username string) error
	FindLink(ctx context.Context, link string) (models.LinkDataDB, error)
	CountLinks(ctx context.Context, username string) (models.LinksAmount, error)
	GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDB, error)
}

// subRepository Интерфейс к слою репозитория подписок Redis
type subRepository interface {
	FindSubscribe(ctx context.Context, username string) (time.Duration, bool)
	AddSubRedis(ctx context.Context, username string, exp time.Duration) error
}

// manager Интерфейс к планировщику задач
type manager interface {
	CleanUnsubscribeSchedule(ctx context.Context, sub models.CurrentSub, username string) error
	CleaningExpLinkSchedule(ctx context.Context, link, username string, exp time.Duration) error
	RemoveCleanSchedule(ctx context.Context, username string)
}
