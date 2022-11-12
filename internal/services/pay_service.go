package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"short_url/internal/models"
	"sync"
	"time"

	"github.com/google/uuid"
)

// userRepository
type userRepository interface {
	AddSubPSQL(ctx context.Context, user models.UserDB) error
	FindByUsername(ctx context.Context, username string) (models.UserDB, error)
}

// subRepository
type subRepository interface {
	AddSubRedis(ctx context.Context, username string, exp time.Duration) error
}

// PayServiceConfig Конфиг для PayService
type PayServiceConfig struct {
	SubRepo		subRepository
	UserRepo	userRepository
	Key			string
	Prices		models.SubPrice
}

// PayService Осуществляет управление оплатой подписок
type PayService struct {
	key			string
	subRepo		subRepository
	userRepo	userRepository
	prices		models.SubPrice
	billIds		map[string]models.SubInfo
	client		*http.Client
	mux			sync.RWMutex
}

// NewPayService Фабрика для PayService
func NewPayService(c *PayServiceConfig) *PayService {
	return &PayService{
		key:		c.Key,
		subRepo:	c.SubRepo,
		userRepo:	c.UserRepo,
		billIds:	make(map[string]models.SubInfo),
		client:		&http.Client{},
	}
}

//
type billReq struct {
	Amount		map[string]any
	Comment		string
	Exp			string
	Customer	struct{}
	Custom		struct{}
}

// Делает запрос на сервер qiwi для создания счета
func (s *PayService) makeBillRequest(amount float64) ([]byte, error) {
	// Создаем уникальную ссылку на счет
	uid := uuid.NewString()

	// Парсим дату для обозначения срока действия счета
	t := time.Now()
	t = t.Add(time.Hour*24)

	// Собираем тело запроса
	BReq := billReq{
		Amount: map[string]any{
			"currency":"RUB",
			"value":amount,
		},
		Comment: "Спасибо что пользуетесь нашим сервисом",
		Exp: t.Format("2006-01-01T00:00:00+01:00"),
		Customer: struct{}{},
		Custom: struct{}{},
	}

	// Парсим тело в формат json
	req, err := json.Marshal(BReq)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(req)

	// Создаем запрос, прикрепляем заголовки и тело
	r, err := http.NewRequest("PUT", "http://api.qiwi.com/partner/bill/v1/bills/"+uid, bodyReader)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", s.key)

	// Отправляем запрос и получаем ответ
	resp, err := s.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем тело ответа в буфер, возвращаем в виде байтов
	return ioutil.ReadAll(resp.Body)
}

// makeRequest Делает запрос на сервер qiwi для проверки статуса счета
func (s *PayService) makeCheckRequest(bill string) ([]byte, error) {
	// Создаем запрос и прикрепляем заголовки
	req, err := http.NewRequest("GET", "https://api.qiwi.com/partner/bill/v1/bills/"+bill, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", s.key)

	// Отправляем запрос и получаем ответ
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем тело ответа в буфер, возвращаем в виде байтов
	return ioutil.ReadAll(resp.Body)
}

// SaveBillId Добавляет Id счета оплаты в кэш
func (s *PayService) SaveBillId(billId string, info models.SubInfo) {
	s.mux.Lock()
	s.billIds[billId] = info
	s.mux.Unlock()
}

// DeleteBillId Удаляет Id счета оплаты из кэша
func (s *PayService) DeleteBillId(billId string) {
	s.mux.Lock()
	delete(s.billIds, billId)
	s.mux.Unlock()
}

// 
func (s *PayService) CalculateSub(amount float64, username string) (models.SubInfo, error) {
	// 
	result := models.SubInfo{Username: username}
	day := time.Hour * 24
	var err error

	// 
	switch amount {
	case s.prices.Weekly:
		result.Exp = day * 7
	case s.prices.Monthly:
		result.Exp = day * 31
	case s.prices.Yearly:
		result.Exp = day * 365
	default:
		err = errors.New("unknown subscribe duration")
	}

	return result, err
}

// QiwiCheck Проверяет в цикле статус оплаты счетов из кэша
func (s *PayService) QiwiCheck(ctx context.Context) chan struct{} {
	// Сигнал остановки
	doneChannel := make(chan struct{}, 1)

	// Тикер (интервал)
	ticker := time.NewTicker(time.Minute * 5)

	// Запускаем пуллинг запросов к серверу qiwi
	go func(doneChannel chan struct{}, ticker *time.Ticker) {

		for {
			select {
			case <-ticker.C:
				// Итерируемся по Id счетов
				s.mux.RLock()
				for bill, subInfo := range s.billIds {
					// Определяем структуру для хранения ответа
					jsonResp := make(map[string]any)

					// Отправляем запрос, получаем ответ
					resp, err := s.makeCheckRequest(bill)
					if err != nil {
						break // log
					}
		
					// Парсим ответ в структуру
					err = json.Unmarshal(resp, &jsonResp)
					if err != nil {
						break // log
					}
		
					// Проверяем поле статуса счета в структуре ответа
					status := jsonResp["status"].(map[string]string)
					if status["value"] != "WAITING" {
						s.mux.RUnlock()
						s.DeleteBillId(bill)
						s.mux.RLock()
						if status["value"] == "PAID" {
							// Ищем пользователя по username
							user, err := s.userRepo.FindByUsername(ctx , subInfo.Username)
							if err != nil {
								break // log
							}
							
							// Меняем пользователю статус подписки в основной таблице
							err = s.userRepo.AddSubPSQL(ctx, user)
							if err != nil {
								break // log
							}

							// Меняем пользователю статус подписки в таблице таймера
							err = s.subRepo.AddSubRedis(ctx, subInfo.Username, subInfo.Exp)
							if err != nil {
								break // log
							}
						}
					}
				}
				s.mux.RUnlock()
			case <-doneChannel:
				// Останавливаем тикер
				ticker.Stop()


				return
			}
		}
	}(doneChannel, ticker)
	
	return doneChannel
}
