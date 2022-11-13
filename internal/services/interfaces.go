package services

import (
	"time"
	"context"
	"short_url/internal/models"

	//cron "github.com/robfig/cron/v3"
)

// authRepository Интерфейс к репозиторию хранения данных пользователей
type authRepository interface {
	FindByUsername(ctx context.Context, name string) (models.UserDB, error)
	CreateUser(ctx context.Context, user models.UserDB) error
}

// linkRepository Интерфейс к репозиторию управления ссылками
type linkRepository interface {
	CreateLink(ctx context.Context, link, username, fullUrl string, exp time.Duration, custom bool) (models.LinkDataDB, error)
	DeleteExpLink(ctx context.Context, link, username string) error
	DeleteLink(ctx context.Context, link, username string) error
	FindLink(ctx context.Context, link string) (models.LinkDataDB, error)
	CountLinks(ctx context.Context, username string) (models.LinksAmount, error)
	GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDB, error)
}

// subRepository Интерфейс к слою репозитория подписок Redis
type subRepository interface {
	AddSubRedis(ctx context.Context, username string, exp time.Duration) error
	FindByUsername(ctx context.Context, username string) (time.Duration, bool)
}

type cronCache interface {
	CleanUnsubscribeSchedule(ctx context.Context, sub models.CurrentSub, username string)
	CleaningExpLinkSchedule(ctx context.Context, link, username string, exp time.Duration)
	RemoveCleanSchedule(username string)
}
