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
	cron "github.com/robfig/cron/v3"
)

// QiwiServiceConfig Конфиг для QiwiService
type QiwiServiceConfig struct {
	Key			string
	SubRepo		subRepository
	AuthRepo	authRepository
	CronCache	cronCache
	Prices		models.SubPrice
	Scheduler	*cron.Cron
}

// QiwiService Осуществляет управление оплатой подписок через QIWI
type QiwiService struct {
	key				string
	subRepo			subRepository
	authRepo		authRepository
	cronCache		cronCache
	prices			models.SubPrice
	scheduler		*cron.Cron
	billIds			map[string]models.SubInfo
	subs			map[string]models.CurrentSub
	client			*http.Client
	mux				sync.RWMutex
}

// NewQiwiService Фабрика для QiwiService
func NewQiwiService(c *QiwiServiceConfig) *QiwiService {
	return &QiwiService{
		key:			c.Key,
		prices:			c.Prices,
		subRepo:		c.SubRepo,
		authRepo:		c.AuthRepo,
		cronCache:		c.CronCache,
		billIds:		make(map[string]models.SubInfo),
		subs:			make(map[string]models.CurrentSub),
		client:			&http.Client{},
	}
}

// billReq Структура запроса для создания счета
type billReq struct {
	Amount		map[string]any
	Comment		string
	Exp			string
	Customer	map[string]any
	Custom		map[string]any
}

// makeBillRequest Делает запрос на сервер qiwi для создания счета
func (s *QiwiService) makeBillRequest(amount float64) (int, string, []byte, error) {
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
		Customer: map[string]any{},
		Custom: map[string]any{},
	}

	// Парсим тело в формат json
	req, err := json.Marshal(BReq)
	if err != nil {
		return 0, "", nil, err
	}
	bodyReader := bytes.NewReader(req)

	// Создаем запрос, прикрепляем заголовки и тело
	r, err := http.NewRequest("PUT", "http://api.qiwi.com/partner/bill/v1/bills/"+uid, bodyReader)
	if err != nil {
		return 0, "", nil, err
	}
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", s.key)

	// Отправляем запрос и получаем ответ
	resp, err := s.client.Do(r)
	if err != nil {
		return 0, "", nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	// Читаем тело ответа в буфер, возвращаем в виде байтов
	return resp.StatusCode, uid, body, err
}

// CalculateSub Рассчитывает срок подписки для пользователя исходя из стоимости
func (s *QiwiService) calculateSub(amount float64, username string) (models.SubInfo, error) {
	// Инициализируем переменные
	result := models.SubInfo{Username: username}
	day := time.Hour * 24
	var err error

	// Выбираем срок подписки
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

// BillRequest
func (s *QiwiService) BillRequest(ctx context.Context, amo float64, username string) (string, error) {
	//
	code, uid, jsonResp, err := s.makeBillRequest(amo)
	if err != nil {
		// log
		return "", err
	}

	//
	if code != http.StatusOK {
		// log
		return "", err
	}

	//
	info, err := s.calculateSub(amo, username)
	if err != nil {
		// log
		return "", err
	}

	//
	err = s.saveBillId(uid, info)
	if err != nil {
		// log
		return "", err
	}

	// Парсим ответ в структуру
	resp := make(map[string]any)
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		// log
		return "", err
	}

	// Извлекаем URL оплаты счета для отправки клиенту
	url, ok := resp["payUrl"].(string)
	if !ok {
		// log
		return "", err
	}

	return url, nil
}

// makeCheckRequest Делает запрос на сервер qiwi для проверки статуса счета
func (s *QiwiService) makeCheckRequest(bill string) ([]byte, error) {
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

// NotifyFromQiwi Обрабатывает уведомление от сервера Qiwi о счете
func (s *QiwiService) NotifyFromQiwi(ctx context.Context, status, bill string) error {
	// Если счет оплачен, добавляем пользователю подписку
	if status == "PAID" {
		// Получаем информацию о клиенте по счету
		info, ok := s.getSubInfo(bill)
		if !ok {
			// log
			return errors.New("")
		}

		//
		exp, ok := s.subRepo.FindByUsername(ctx, info.Username)
		if ok {
			s.cronCache.RemoveCleanSchedule(info.Username)
		}

		//
		sub := models.CurrentSub{
			Exp: info.Exp + exp,
		}

		//
		s.cronCache.CleanUnsubscribeSchedule(ctx, sub, info.Username)

		// Меняем пользователю статус подписки во временном хранилище подписчиков
		err := s.subRepo.AddSubRedis(ctx, info.Username, info.Exp)
		if err != nil {
			// log
			return err
		}
	}

	// Удаляем счет, так как он больше не WAITING (либо оплачен, либо просрочен, либо отменен)
	ok := s.deleteBillId(bill)
	if !ok {
		// log
		return errors.New("")
	}

	return nil
}

// QiwiCheck Проверяет в цикле счета из кэша, находившиеся в статусе WAITING
func (s *QiwiService) QiwiCheck(ctx context.Context) chan struct{} {
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
					resp := make(map[string]any)

					// Отправляем запрос, получаем ответ
					jsonResp, err := s.makeCheckRequest(bill)
					if err != nil {
						break // log
					}
		
					// Парсим ответ в структуру
					err = json.Unmarshal(jsonResp, &resp)
					if err != nil {
						break // log
					}
		
					// Проверяем поле статуса счета в структуре ответа
					status, ok := resp["status"].(map[string]string)
					if !ok {
						// log
					}

					//
					if status["value"] != "WAITING" {
						s.mux.RUnlock()
						ok := s.deleteBillId(bill)
						if !ok {
							// log
						}
						s.mux.RLock()

						//
						if status["value"] == "PAID" {
							//
							exp, ok := s.subRepo.FindByUsername(ctx, subInfo.Username)
							if ok {
								//\\\\\\\\\\\\\\\\\\
							}

							//
							sub := models.CurrentSub{
								Exp: subInfo.Exp + exp,
							}

							//
							s.cronCache.CleanUnsubscribeSchedule(ctx, sub, subInfo.Username)

							// Меняем пользователю статус подписки во временном хранилище подписчиков
							err := s.subRepo.AddSubRedis(ctx, subInfo.Username, subInfo.Exp)
							if err != nil {
								// log
								return 
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

// SaveBillId Добавляет Id счета оплаты в кэш
func (s *QiwiService) saveBillId(billId string, info models.SubInfo) error {
	_, ok := s.getSubInfo(billId)
	if ok {
		return errors.New("billId already exists")
	}
	s.mux.Lock()
	s.billIds[billId] = info
	s.mux.Unlock()
	return nil
}

// DeleteBillId Удаляет Id счета оплаты из кэша
func (s *QiwiService) deleteBillId(billId string) bool {
	_, ok := s.getSubInfo(billId)
	if !ok {
		return false
	}
	s.mux.Lock()
	delete(s.billIds, billId)
	s.mux.Unlock()
	return true
}

// GetSubInfo Получает информацию о пользователе и сроке подписки по номеру счета из кэша
func (s *QiwiService) getSubInfo(billId string) (models.SubInfo, bool) {
	s.mux.RLock()
	info, ok := s.billIds[billId]
	s.mux.RUnlock()
	return info, ok
}

// CleaningUnsubscribeSchedule
// func (s *QiwiService) CleaningUnsubscribeSchedule(ctx context.Context, sub models.CurrentSub, username string) {
// 	//
// 	sub.RemId = make([]cron.EntryID, 0)

// 	//
// 	amo, err := s.linkRepo.CountLinks(ctx, username)
// 	if err != nil {
// 		// log
// 		return 
// 	}

// 	//
// 	var counter, customCounter int

// 	//
// 	defaultLinks := make([]models.LinkDataDB, 0)
// 	customLinks := make([]models.LinkDataDB, 0)

// 	//
// 	links, err := s.linkRepo.GetAllLinks(ctx, username)

// 	//
// 	for k := 0; k < len(links); k++ {
// 		if links[k].Perm {
// 			//
// 			id, err := s.scheduler.AddFunc("////////", func() {
// 				s.linkRepo.DeleteLink(ctx, links[k].Link, username)
// 			})
// 			if err != nil {
// 				// log
// 				return 
// 			}

// 			sub.RemId = append(s.subs[username].RemId, id)
// 		} else {
// 			if links[k].Custom {
// 				customLinks = append(customLinks, links[k])
// 			} else {
// 				defaultLinks = append(defaultLinks, links[k])
// 			}
// 		}
// 	}

// 	//
// 	if amo.Custom > Custom {
// 		customCounter = Custom - amo.Custom
// 	}

// 	//
// 	for k := 0; k < customCounter; k++ {
// 			//
// 			id, err := s.scheduler.AddFunc("////////", func() {
// 				s.linkRepo.DeleteLink(ctx, customLinks[k].Link, username)
// 			})
// 			if err != nil {
// 				// log
// 				return 
// 			}

// 			sub.RemId = append(s.subs[username].RemId, id)
// 	}

// 	//
// 	if len(defaultLinks) + len(customLinks) - customCounter > All {
// 		counter = All - len(defaultLinks) - len(customLinks) + customCounter
// 	}

// 	//
// 	for k := 0; k < counter; k++ {
// 		//
// 		id, err := s.scheduler.AddFunc("////////", func() {
// 			s.linkRepo.DeleteLink(ctx, defaultLinks[k].Link, username)
// 		})
// 		if err != nil {
// 			// log
// 			return 
// 		}

// 		sub.RemId = append(s.subs[username].RemId, id)
// 	}

// 	return
// }
