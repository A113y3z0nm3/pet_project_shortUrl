package services

import (
	"context"
	"errors"
	"math/rand"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// linkRepository Интерфейс к репозиторию управления ссылками
type linkRepository interface {
	CreateLink(ctx context.Context, link, username, fullUrl string, exp time.Duration, custom bool) (models.LinkDataDB, error)
	DeleteLink(ctx context.Context, link, username string) error
	FindLink(ctx context.Context, link string) (models.LinkDataDB, error)
	CountLinks(ctx context.Context, username string) (models.LinksAmount, error)
	GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDB, error)
}

// LinkServiceConfig Конфигурация для LinkService
type LinkServiceConfig struct {
	LinkRepo linkRepository
	Logger   *myLog.Log
}

// LinkService Управляет взаимодействием с ссылками
type LinkService struct {
	linkRepo linkRepository
	logger   *myLog.Log
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
		linkRepo: c.LinkRepo,
		logger:   c.Logger,
	}
}

// CreateQR Создает QR-код по существующей короткой ссылке
func (s *LinkService) CreateQR(ctx context.Context, url, link string) (barcode.Barcode, error) {

	// Находим ссылку в БД
	_, err := s.linkRepo.FindLink(ctx, link)
	if err != nil {
		return nil, err
	}

	// Создаем QR-код на основе короткой ссылки
	qrCode, err := qr.Encode(url, qr.M, qr.Auto)
	if err != nil {
		return nil, err
	}

	// Форматируем QR-код
	qrCode, err = barcode.Scale(qrCode, 256, 256)
	if err != nil {
		return nil, err
	}

	return qrCode, nil
}

// FindLink Находит ссылку и доп. информацию о ней
func (s *LinkService) FindLink(ctx context.Context, link string) (models.LinkDataDTO, error) {

	// Возвращает ссылку из БД
	data, err := s.linkRepo.FindLink(ctx, link)
	if err != nil {
		return models.LinkDataDTO{}, err
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

	// Получаем все ссылки пользователя из БД
	data, err := s.linkRepo.GetAllLinks(ctx, username)
	if err != nil {
		return nil, err
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

	// Удаляем ссылку из БД
	err := s.linkRepo.DeleteLink(ctx, link, username)
	if err != nil {
		return err
	}

	return nil
}

// CreateLink Проверяет выполнение условий сервиса и создает ссылку
func (s *LinkService) CreateLink(ctx context.Context, fullUrl, custom string, exp int, user models.JWTUserInfo) (models.LinkDataDTO, error) {

	// Считаем кол-во ссылок у пользователя
	amo, err := s.linkRepo.CountLinks(ctx, user.Username)
	if err != nil {
		return models.LinkDataDTO{}, err
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
		return models.LinkDataDTO{}, err
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
