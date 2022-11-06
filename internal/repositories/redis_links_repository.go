package repositories

import (
	"context"
	"short_url/internal/models"
	"time"

	"github.com/go-redis/redis/v9"
)

// RedisLinkRepositoryConfig Конфигурация для RedisLinkRepository
type RedisLinkRepositoryConfig struct {
	DB		*redis.Client
}

// RedisLinkRepository Слой для управления запросами к хранилищу ссылок
type RedisLinkRepository struct {
	db		*redis.Client
	pipe 	redis.Pipeliner
}

// Структура для записи в таблицу ссылок
type RedisNote map[string]any

const (
	p	=	"perm"
	c	=	"custom"
	u	=	"url"
)

// NewRedisLinkRepository Конструктор для RedisLinkRepository
func NewRedisLinkRepository(c *RedisLinkRepositoryConfig) *RedisLinkRepository {
	redisRepo := &RedisLinkRepository{
		db:		c.DB,
	}

	// Инициализируем пайплайн (выполняет несколько команд за одну запись)
	redisRepo.pipe = redisRepo.db.Pipeline()

	return redisRepo
}

// CreateLinkTimed Создает ссылку в таблицах
func (r *RedisLinkRepository) CreateLink(ctx context.Context, link, username, fullUrl string, exp time.Duration, custom bool) (models.LinkDataDB, error) {
	
	// Определяем, с таймером ли ссылка
	var perm bool
	if exp == 0 {
		perm = true
	}

	// Вставляем ссылку в таблицу пользователя
	_, err := r.pipe.SAdd(ctx, username, link).Result()
	if err != nil {
		return models.LinkDataDB{}, err
	}

	// Вставляем метаданные в таблицу ссылок
	metaLink := "meta-" + link
	_, err = r.pipe.HSet(ctx, metaLink, RedisNote{p:perm, c:custom, u:fullUrl}).Result()
	if err != nil {
		return models.LinkDataDB{}, err
	}

	// Вставляем ссылку в таблицу таймера
	_, err = r.pipe.Set(ctx, link, 1, exp).Result()
	if err != nil {
		return models.LinkDataDB{}, err
	}

	// Маппим данные в результат
	return models.LinkDataDB{
		Link:		link,
		FullURL:	fullUrl,
		ExpTime:	exp,
		Perm:		perm,
		Custom: 	custom,
	}, nil
}

// DeleteLink Удаляет записи из таблицы
func (r *RedisLinkRepository) DeleteLink(ctx context.Context, link, username string) error {

	metaLink := "meta-" + link

	// Удаляем метаданные из таблицы ссылок
	_, err := r.pipe.Del(ctx, metaLink).Result()
	if err != nil {
		return err
	}

	// Удаляем ссылку из таблицы пользователя
	_, err = r.pipe.SRem(ctx, username, link).Result()
	if err != nil {
		return err
	}

	// Удаляем ссылку из таблицы таймера
	_, err = r.pipe.Del(ctx, link).Result()
	if err != nil {
		return err
	}

	return nil
}

// FindLink Находит ссылку и метаданные о ней
func (r *RedisLinkRepository) FindLink(ctx context.Context, link string) (models.LinkDataDB, error) {

	metaLink := "meta-" + link

	// Инициализируем переменные
	result := models.LinkDataDB{}
	var err error

	// Записываем значения и обрабатываем ошибки
	result.Link = link
	data, err := r.pipe.HMGet(ctx, metaLink, p, c, u).Result()
	if err != nil {
		return result, err
	}
	result.Perm = data[0].(bool)
	result.Custom = data[1].(bool)
	result.FullURL = data[2].(string)
	result.ExpTime, err = r.pipe.TTL(ctx, link).Result()
	if err != nil {
		return result, err
	}

	return result, nil
}

// CountLinks Считает кол-во ссылок на аккаунте пользователя
func (r *RedisLinkRepository) CountLinks(ctx context.Context, username string) (models.LinksAmount, error) {

	// Инициализируем структуру ответа
	result := models.LinksAmount{}

	// Получаем все ссылки из таблицы пользователя
	data, err := r.db.SMembers(ctx, username).Result()
	if err != nil {
		return result, err
	}

	// Инициализируем счетчики особых ссылок
	var perm, custom int

	// Итерируемся по таблице метаданных, записывая кол-во особых ссылок пользователя
	for _, k := range data {

		metaLink := "meta-" + k

		cell, err := r.db.HMGet(ctx, metaLink, p, c).Result()
		if err != nil {
			return result, err
		}

		if cell[0] == true {
			perm += 1
		}
		if cell[1] == true {
			custom += 1
		}
	}

	// Маппим значения в ответ
	result.All = len(data)
	result.Custom = custom
	result.Perm = perm

	return result, nil
}

// GetAllLinks Получает все ссылки пользователя из таблицы ссылок
func (r *RedisLinkRepository) GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDB, error) {

	// Получаем все ссылки из таблицы пользователя
	data, err := r.db.SMembers(ctx, username).Result()
	if err != nil {
		return []models.LinkDataDB{}, err
	}

	// Инициализируем структуру ответа
	result := []models.LinkDataDB{}

	// Итерируемся по таблицам метаданных и ссылок, записывая данные в ответ
	for q, k := range data {

		metaLink := "meta-" + k

		cell, err := r.db.HMGet(ctx, metaLink, p, c, u).Result()
		if err != nil {
			return result, err
		}

		result[q].Link = k
		result[q].FullURL = cell[2].(string)
		result[q].Perm = cell[0].(bool)
		result[q].Custom = cell[1].(bool)
		result[q].ExpTime, err = r.pipe.TTL(ctx, k).Result()
		if err != nil {
			return result, err
		}
	}

	return result, nil
}
