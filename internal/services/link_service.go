package services

import (
	"bytes"
	"context"
	"errors"
	"image/png"
	"math/rand"
	"short_url/internal/models"
	log "short_url/pkg/logger"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/go-redis/redis/v9"
)

// LinkServiceConfig Конфигурация для LinkService
type LinkServiceConfig struct {
	LinkRepo	linkRepository
	Manager	manager
	Logger		*log.Log
}

// LinkService Управляет взаимодействием с ссылками
type LinkService struct {
	linkRepo 	linkRepository
	manager	manager
	logger   	*log.Log
}

const (
	DefaultLifeTime	= 12 * time.Hour	// Время жизни ссылок по умолчанию

	All    			= 50				// Лимит кол-ва всех ссылок для обычного пользователя
	SubAll 			= 100				// Лимит кол-ва всех ссылок для подписчика

	Custom    		= 15				// Лимит кол-ва кастомных ссылок для обычного пользователя
	SubCustom 		= 30				// Лимит кол-ва кастомных ссылок для подписчика

	SubPerm			= 10				// Лимит кол-ва ссылок с безграничным сроком действия для подписчика
)

// NewLinkService Конструктор для ManageService
func NewLinkService(c *LinkServiceConfig) *LinkService {
	return &LinkService{
		linkRepo:	c.LinkRepo,
		manager:	c.Manager,
		logger:		c.Logger,
	}
}

// CreateQR Создает QR-код по существующей короткой ссылке и возвращает его в виде байтов
func (s *LinkService) CreateQR(ctx context.Context, url, link string) (*bytes.Buffer, error) {
	ctx = log.ContextWithSpan(ctx, "CreateQR")
	l := s.logger.WithContext(ctx)

	l.Debug("CreateQR() started")
	defer l.Debug("CreateQR() done")

	// Находим ссылку в БД
	_, err := s.linkRepo.FindLink(ctx, link)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("link not found")
		} else {
			l.Errorf("Unable to find link in Redis. Error: %s", err)
			return nil, err
		}
	}

	// Создаем QR-код на основе короткой ссылки
	qrCode, err := qr.Encode(url, qr.M, qr.Auto)
	if err != nil {
		l.Errorf("Unable to encode link to QR. Error: %s", err)
		return nil, err
	}

	// Форматируем QR-код
	qrCode, err = barcode.Scale(qrCode, 256, 256)
	if err != nil {
		l.Errorf("Unable to format QR. Error: %s", err)
		return nil, err
	}

	result := bytes.NewBuffer([]byte{})

	err = png.Encode(result, qrCode)

	return result, nil
}

// FindLink Находит ссылку и доп. информацию о ней
func (s *LinkService) FindLink(ctx context.Context, link string) (models.LinkDataDTO, error) {
	ctx = log.ContextWithSpan(ctx, "FindLink")
	l := s.logger.WithContext(ctx)

	l.Debug("FindLink() started")
	defer l.Debug("FindLink() done")

	// Возвращает ссылку из БД
	data, err := s.linkRepo.FindLink(ctx, link)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.LinkDataDTO{}, errors.New("link not found")
		} else {
			l.Errorf("Unable to find link in Redis. Error: %s", err)
			return models.LinkDataDTO{}, err
		}
	}

	// Маппим данные в ответ
	result := models.LinkDataDTO{
		Link:    data.Link,
		FullURL: data.FullURL,
		ExpTime: int(data.ExpTime),
	}

	return result, nil
}

// GetAllLinks Возвращает все ссылки пользователя
func (s *LinkService) GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDTO, error) {
	ctx = log.ContextWithSpan(ctx, "GetAllLinks")
	l := s.logger.WithContext(ctx)

	l.Debug("GetAllLinks() started")
	defer l.Debug("GetAllLinks() done")

	// Получаем все ссылки пользователя из БД
	data, err := s.linkRepo.GetAllLinks(ctx, username)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("links not found")
		} else {
			l.Errorf("Unable to get all links from Redis. Error: %s", err)
			return nil, err
		}
	}

	// Маппим данные в ответ
	result := make([]models.LinkDataDTO, len(data))
	for k, d := range data {
		result[k].Link = d.Link
		result[k].FullURL = d.FullURL
		result[k].ExpTime = int(d.ExpTime)
	}

	return result, nil
}

// DeleteLink Удаляет ссылку
func (s *LinkService) DeleteLink(ctx context.Context, username, link string) error {
	ctx = log.ContextWithSpan(ctx, "DeleteLink")
	l := s.logger.WithContext(ctx)

	l.Debug("DeleteLink() started")
	defer l.Debug("DeleteLink() done")

	// Удаляем ссылку из БД
	err := s.linkRepo.DeleteLink(ctx, link, username)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("link not found")
		} else {
			l.Errorf("Unable to get all links from Redis. Error: %s", err)
			return err
		}
	}

	return nil
}

// CreateLink Проверяет выполнение условий сервиса и создает ссылку
func (s *LinkService) CreateLink(ctx context.Context, fullUrl, custom string, exp int, user models.JWTUserInfo) (models.LinkDataDTO, error) {
	ctx = log.ContextWithSpan(ctx, "CreateLink")
	l := s.logger.WithContext(ctx)

	l.Debug("CreateLink() started")
	defer l.Debug("CreateLink() done")

	// Считаем кол-во ссылок у пользователя
	amo, err := s.linkRepo.CountLinks(ctx, user.Username)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			l.Errorf("Unable to get all links from Redis. Error: %s", err)
			return models.LinkDataDTO{}, err
		}
	}

	// Вводим ограничения сервиса
	if user.Subscribe == 2 {
		if exp == 0 {
			return models.LinkDataDTO{}, errors.New("need subscribe")
		}
		if custom != "" {
			if amo.Custom == Custom {
				return models.LinkDataDTO{}, errors.New("limit exceeded")
			}
		}
		if amo.All == All {
			return models.LinkDataDTO{}, errors.New("limit exceeded")
		}
	} else {
		if exp == 0 {
			if amo.Perm == SubPerm {
				return models.LinkDataDTO{}, errors.New("limit exceeded")
			}
		}
		if custom != "" {
			if amo.Custom == SubCustom {
				return models.LinkDataDTO{}, errors.New("limit exceeded")
			}
		}
		if amo.All == SubAll {
			return models.LinkDataDTO{}, errors.New("limit exceeded")
		}
	}

	// Определяем, кастомная ли ссылка
	var link string
	var isCustom bool
	if custom != "" {
		link = custom
		isCustom = true
	} else {
		link = randLink()
		isCustom = false
	}

	// Добавляем ссылку в БД
	data, err := s.linkRepo.CreateLink(ctx, link, user.Username, fullUrl, time.Duration(exp), isCustom)
	if err != nil {
		l.Errorf("Unable to create link data in Redis. Error: %s", err)
		return models.LinkDataDTO{}, err
	}

	// Если параметр срока действия ссылки обозначен, планируем задачу на удаление по истечению срока
	if exp != 0 {
		if err = s.manager.CleaningExpLinkSchedule(ctx, link, user.Username, time.Duration(exp)); err != nil {
			l.Errorf("Unable to schedule cleaning. Error: %s", err)
			return models.LinkDataDTO{}, err
		}
	}

	// Маппим данные в ответ
	result := models.LinkDataDTO{
		Link:    data.Link,
		FullURL: data.FullURL,
		ExpTime: int(data.ExpTime),
	}

	return result, nil
}

// randLink Генератор рандомной ссылки
func randLink() string {

	rand.Seed(time.Now().UnixNano())

	alphabete := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" + "abcdefghijklmnopqrstuvwxyz"
	length := 8

	buf := make([]byte, length)
	for i := 2; i < length; i++ {
		buf[i] = alphabete[rand.Intn(len(alphabete))]
	}

	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})

	return string(buf)
}
