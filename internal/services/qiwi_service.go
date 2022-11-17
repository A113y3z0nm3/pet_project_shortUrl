package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"short_url/internal/models"
	log "short_url/pkg/logger"
	"sync"
	"time"

	"github.com/google/uuid"
)

// QiwiServiceConfig Конфиг для QiwiService
type QiwiServiceConfig struct {
	Key			string
	SubRepo		subRepository
	AuthRepo	authRepository
	Manager		manager
	Prices		models.ConfigPrice
	Logger		*log.Log
}

// QiwiService Осуществляет управление оплатой подписок через QIWI
type QiwiService struct {
	key				string
	subRepo			subRepository
	authRepo		authRepository
	manager			manager
	prices			models.ConfigPrice
	billIds			map[string]models.SubInfo
	subs			map[string]models.CurrentSub
	client			*http.Client
	mux				sync.RWMutex
	logger			*log.Log
}

// NewQiwiService Фабрика для QiwiService
func NewQiwiService(c *QiwiServiceConfig) *QiwiService {
	return &QiwiService{
		key:			c.Key,
		prices:			c.Prices,
		subRepo:		c.SubRepo,
		authRepo:		c.AuthRepo,
		billIds:		make(map[string]models.SubInfo),
		subs:			make(map[string]models.CurrentSub),
		client:			&http.Client{},
		logger: 		c.Logger,
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
func (s *QiwiService) makeBillRequest(ctx context.Context, amount float64) (int, string, []byte, error) {
	ctx = log.ContextWithSpan(ctx, "makeBillRequest")
	l := s.logger.WithContext(ctx)

	l.Debug("makeBillRequest() started")
	defer l.Debug("makeBillRequest() done")

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
		l.Errorf("Unable to marshal request body. Error: %s", err)
		return 0, "", nil, err
	}
	bodyReader := bytes.NewReader(req)

	// Создаем запрос, прикрепляем заголовки и тело
	r, err := http.NewRequest("PUT", "http://api.qiwi.com/partner/bill/v1/bills/"+uid, bodyReader)
	if err != nil {
		l.Errorf("Unable to build PUT request to QIWI. Error: %s", err)
		return 0, "", nil, err
	}
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", s.key)

	// Отправляем запрос и получаем ответ
	resp, err := s.client.Do(r)
	if err != nil {
		l.Errorf("Unable to push PUT request to QIWI. Error: %s", err)
		return 0, "", nil, err
	}
	defer resp.Body.Close()

	// Читаем тело ответа в буфер, возвращаем в виде байтов
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("Unable to read response body. Error: %s", err)
		return 0, "", nil, err
	}

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

// BillRequest Создает счет на оплату в QIWI и парсит ссылку на оплату из тела ответа
func (s *QiwiService) BillRequest(ctx context.Context, amo float64, username string) (string, error) {
	ctx = log.ContextWithSpan(ctx, "BillRequest")
	l := s.logger.WithContext(ctx)

	l.Debug("BillRequest() started")
	defer l.Debug("BillRequest() done")

	// Отправляем запрос
	code, uid, jsonResp, err := s.makeBillRequest(ctx, amo)
	if err != nil {
		l.Errorf("Unable to make bill request. Error: %s", err)
		return "", err
	}

	// Если код не успешный, сбрасываем операцию
	if code != http.StatusOK {
		return "", l.RErrorf("Error: HTTP Response code (%s) not equal 200", code)
	}

	// Считаем срок подписки
	info, err := s.calculateSub(amo, username)
	if err != nil {
		l.Errorf("Unable to calculate subscribe duration. Error: %s", err)
		return "", err
	}

	// Сохраняем номер счета для мониторинга его статуса
	err = s.saveBillId(uid, info)
	if err != nil {
		l.Errorf("Unable to save billId to cache. Error: %s", err)
		return "", err
	}

	// Парсим ответ в структуру
	resp := make(map[string]any)
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		l.Errorf("Unable to unmarshal response body. Error: %s", err)
		return "", err
	}

	// Извлекаем URL оплаты счета для отправки клиенту
	url, ok := resp["payUrl"].(string)
	if !ok {
		return "", l.RError("Error: Unable to get pay url from response body")
	}

	return url, nil
}

// makeCheckRequest Делает запрос на сервер qiwi для проверки статуса счета
func (s *QiwiService) makeCheckRequest(ctx context.Context, bill string) ([]byte, error) {
	ctx = log.ContextWithSpan(ctx, "makeCheckRequest")
	l := s.logger.WithContext(ctx)

	l.Debug("makeCheckRequest() started")
	defer l.Debug("makeCheckRequest() done")

	// Создаем запрос и прикрепляем заголовки
	req, err := http.NewRequest("GET", "https://api.qiwi.com/partner/bill/v1/bills/"+bill, nil)
	if err != nil {
		l.Errorf("Unable to build GET request to QIWI", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", s.key)

	// Отправляем запрос и получаем ответ
	resp, err := s.client.Do(req)
	if err != nil {
		l.Errorf("Unable to push GET request to QIWI", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем тело ответа в буфер, возвращаем в виде байтов
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("Unable to read response body. Error: %s", err)
		return body, err
	}

	return body, nil
}

// AddSubscribe Находит пользователя в репозитории подписок, если найден - обновляет подписку, если нет - добавляет
func (s *QiwiService) AddSubscribe(ctx context.Context, info models.SubInfo) error {
	ctx = log.ContextWithSpan(ctx, "AddSubscribe")
	l := s.logger.WithContext(ctx)

	l.Debug("AddSubscribe() started")
	defer l.Debug("AddSubscribe() done")

	// Модель текущей подписки
	sub := models.CurrentSub{}

	//
	exp, ok := s.subRepo.FindSubscribe(ctx, info.Username)
	if ok {
		// Отменяем назначенные операции по чистке (подписка продлена)
		s.manager.RemoveCleanSchedule(ctx, info.Username)

		// Увеличиваем срок подписки
		sub.Exp = info.Exp + exp
	} else {
		// Задаем срок подписки
		sub.Exp = info.Exp
	}

	// Планируем чистку по окончанию подписки
	err := s.manager.CleanUnsubscribeSchedule(ctx, sub, info.Username)
	if err != nil {
		l.Errorf("Unable to schedule data cleaning. Error: %s", err)
		return err
	}

	// Меняем пользователю статус подписки во временном хранилище подписчиков
	err = s.subRepo.AddSubRedis(ctx, info.Username, info.Exp)
	if err != nil {
		l.Errorf("Unable to subscribe user. Error: %s", err)
		return err
	}

	return nil
}

// NotifyFromQiwi Обрабатывает уведомление от сервера Qiwi о счете
func (s *QiwiService) NotifyFromQiwi(ctx context.Context, status, bill string) error {
	ctx = log.ContextWithSpan(ctx, "NotifyFromQiwi")
	l := s.logger.WithContext(ctx)

	l.Debug("NotifyFromQiwi() started")
	defer l.Debug("NotifyFromQiwi() done")

	// Если счет оплачен, добавляем пользователю подписку
	if status == "PAID" {
		// Получаем информацию о клиенте по счету
		info, ok := s.getSubInfo(bill)
		if !ok {
			return l.RError("Unable to get information about subscribe")
		}

		// Оформляем подписку
		err := s.AddSubscribe(ctx, info)
		if err != nil {
			l.Errorf("Unable to add subscribe to user. Error: %s", err)
			return err
		}
	}

	// Удаляем счет, так как он больше не WAITING (либо оплачен, либо просрочен, либо отменен)
	err := s.deleteBillId(bill)
	if err != nil {
		l.Errorf("Unable to delete billId from cache. Error: %s", err)
	}

	return nil
}

// QiwiCheck Проходит по счетам в кэше и выполняет операции по ним
func (s *QiwiService) QiwiCheck(ctx context.Context) error {
	ctx = log.ContextWithSpan(ctx, "QiwiCheck")
	l := s.logger.WithContext(ctx)

	l.Debug("QiwiCheck() started")
	defer l.Debug("QiwiCheck() done")

	// Слайс с номерами счетов для удаления из кэша (в статусе != WAITING)
	billToDelete := make([]string, 0)

	s.mux.RLock()
	for bill, subInfo := range s.billIds {
		// Определяем структуру для хранения ответа
		resp := make(map[string]any)

		// Отправляем запрос, получаем ответ
		jsonResp, err := s.makeCheckRequest(ctx, bill)
		if err != nil {
			l.Errorf("Unable to make GET request to QIWI. Error: %s", err)
			return err // log
		}

		// Парсим ответ в структуру
		err = json.Unmarshal(jsonResp, &resp)
		if err != nil {
			l.Errorf("Unable to unmarshal response body. Error: %s", err)
			return err
		}

		// Проверяем поле статуса счета в структуре ответа
		status, ok := resp["status"].(map[string]string)
		if !ok {
			return l.RError("Unable to get bill status from response body")
		}

		// Если статус счета уже не в ожидании, можно удалять его из кэша
		if status["value"] != "WAITING" {
			
			// Добавляем счет на удаление
			billToDelete = append(billToDelete, bill)

			// Если статус платежа оплачен, присваиваем пользователю подписку
			if status["value"] == "PAID" {
				// Оформляем подписку
				err := s.AddSubscribe(ctx, subInfo)
				if err != nil {
					l.Errorf("Unable to add subscribe to user. Error: %s", err)
					return err
				}
			}
		}
	}
	s.mux.RUnlock()

	// Удаляем из кэша каждый счет, помеченый на удаление
	for k := 0; k < len(billToDelete); k++ {
		err := s.deleteBillId(billToDelete[k])
		if err != nil {
			l.Errorf("Unable to delete useless billId. Error: %s", err)
		}
	}

	return nil
}

// QiwiCheck Проверяет в цикле счета из кэша, находившиеся в статусе WAITING
func (s *QiwiService) QiwiCheckCycle(ctx context.Context) chan struct{} {
	ctx = log.ContextWithSpan(ctx, "QiwiCheckCycle")
	l := s.logger.WithContext(ctx)

	// Канал сигнала остановки
	doneChannel := make(chan struct{}, 1)

	// Тикер (интервал)
	ticker := time.NewTicker(time.Minute * 5)

	// Запускаем пуллинг запросов к серверу qiwi
	go func(doneChannel chan struct{}, ticker *time.Ticker) {
		l.Info("start qiwi check")
		for {
			select {
			case <-ticker.C:
				// Итерируемся по Id счетов
				s.QiwiCheck(ctx)
				
			case <-doneChannel:
				// Останавливаем тикер
				ticker.Stop()

				l.Info("end qiwi check")

				return
			}
		}
	}(doneChannel, ticker)
	
	return doneChannel
}

// saveBillId Добавляет Id счета оплаты в кэш
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

// deleteBillId Удаляет Id счета оплаты из кэша
func (s *QiwiService) deleteBillId(billId string) error {
	_, ok := s.getSubInfo(billId)
	if !ok {
		return errors.New("billId not found")
	}
	s.mux.Lock()
	delete(s.billIds, billId)
	s.mux.Unlock()
	return nil
}

// getSubInfo Получает информацию о пользователе и сроке подписки по номеру счета из кэша
func (s *QiwiService) getSubInfo(billId string) (models.SubInfo, bool) {
	s.mux.RLock()
	info, ok := s.billIds[billId]
	s.mux.RUnlock()
	return info, ok
}
