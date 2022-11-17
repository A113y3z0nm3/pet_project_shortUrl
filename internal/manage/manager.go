package manager

import (
	"context"
	"fmt"
	"short_url/internal/models"
	log "short_url/pkg/logger"
	"sync"
	"time"

	cron "github.com/robfig/cron/v3"
)

const (
	All    			= 50				// Лимит кол-ва всех ссылок для обычного пользователя
	//SubAll 			= 100				// Лимит кол-ва всех ссылок для подписчика

	Custom    		= 15				// Лимит кол-ва кастомных ссылок для обычного пользователя
	//SubCustom 		= 30				// Лимит кол-ва кастомных ссылок для подписчика

	//SubPerm			= 10				// Лимит кол-ва ссылок с безграничным сроком действия для подписчика
)

// Нулевое время (для сравнения полей задач)
var TimeNil time.Time

// linkRepository Интерфейс к репозиторию управления ссылками
type linkRepository interface {
	DeleteExpLink(ctx context.Context, link, username string) error
	DeleteLink(ctx context.Context, link, username string) error
	CountLinks(ctx context.Context, username string) (models.LinksAmount, error)
	GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDB, error)
}

// ManagerConfig Конфиг для Manager
type ManagerConfig struct {
	LinkRepo		linkRepository
	Scheduler		*cron.Cron
	Logger			*log.Log
}

// Manager Вспомогательный слой между repositories и services для планировки задач
type Manager struct {
	linkRepo		linkRepository
	scheduler		*cron.Cron
	subs			map[string]models.CurrentSub
	links			[]cron.EntryID
	mux				sync.RWMutex
	logger			*log.Log
}

// NewManager Конструктор для Manager
func NewManager(conf *ManagerConfig) *Manager {
	return &Manager{
		linkRepo: conf.LinkRepo,
		scheduler: conf.Scheduler,
		subs: make(map[string]models.CurrentSub),
		logger: conf.Logger,
		links: make([]cron.EntryID, 0),
	}
}

// CheckJobs Проверяет задачи и контролирует чтобы они выполнялись один раз, чистит буфер планировщика
func (c *Manager) CheckJobs(ctx context.Context) {
	ctx = log.ContextWithSpan(ctx, "CheckJobs")
	l := c.logger.WithContext(ctx)

	l.Debug("CheckJobs() started")
	defer l.Debug("CheckJobs() done")

	// Итерируемся по кэшу ссылок
	for k := 0; k < len(c.links); k++ {
		id := c.links[k]

		// Получаем сведения о задаче
		job := c.scheduler.Entry(id)

		// Если существует предыдущее время выполнения, удаляем задачу
		if job.Prev != TimeNil {
			c.scheduler.Remove(id)
		}
	}

	// Итерируемся по кэшу подписок
	c.mux.RLock()
	for _, sub := range c.subs {
		// Итерируемся по задачам пользователя
		for k := 0; k < len(sub.RemId); k++ {
			id := sub.RemId[k]
			
			// Получаем сведения о задаче
			job := c.scheduler.Entry(id)

			// Если существует предыдущее время выполнения, удаляем задачу
			if job.Prev != TimeNil {
				c.scheduler.Remove(id)
			}
		}
	}
	c.mux.RUnlock()

	return
}

// SchedChecker Проверяет в цикле задачи на предмет их единовременного выполнения
func (c *Manager) SchedChecker(ctx context.Context) chan struct{} {
	ctx = log.ContextWithSpan(ctx, "SchedChecker")
	l := c.logger.WithContext(ctx)

	// Канал сигнала остановки
	doneChannel := make(chan struct{}, 1)

	// Тикер (интервал)
	ticker := time.NewTicker(time.Hour * 12)

	// Запускаем пуллинг запросов к серверу qiwi
	go func(doneChannel chan struct{}, ticker *time.Ticker) {
		l.Info("start sched check")
		for {
			select {
			case <-ticker.C:
				// Итерируемся по Id счетов
				c.CheckJobs(ctx)
			case <-doneChannel:
				// Останавливаем тикер
				ticker.Stop()

				l.Info("end sched check")

				return
			}
		}
	}(doneChannel, ticker)
	
	return doneChannel
}

// formatSched Рассчитывает время удаления и переводит его в формат планировщика
func formatSched(exp time.Duration) string {
	// Добавляем длительность к настоящему времени
	date := time.Now().Add(exp)

	// Парсим составляющие даты
	mi := date.Minute()
	h := date.Hour()
	d := date.Day()
	mo := int(date.Month())

	// Вставляем значения в шаблон планировщика
	result := fmt.Sprintf("%v %v %v %v *", mi, h, d, mo)
	return result
}

// CleaningExpLinkSchedule Планирует удаление ссылок из типов Redis, которым нельзя указать срок действия (множество, хеш)
func (c *Manager) CleaningExpLinkSchedule(ctx context.Context, link, username string, exp time.Duration) error {
	ctx = log.ContextWithSpan(ctx, "CleaningExpLinkSchedule")
	l := c.logger.WithContext(ctx)

	l.Debug("CleaningExpLinkSchedule() started")
	defer l.Debug("CleaningExpLinkSchedule() done")

	// Задаем задачу планировщику и получаем ее номер
	id, err := c.scheduler.AddFunc(formatSched(exp), func(){
		c.linkRepo.DeleteExpLink(ctx, link, username)
	})
	if err != nil {
		l.Errorf("Unable to add scheduler job. Error: %s", err)
		return err
	}

	// Добавляем номер в кэш для будущего удаления
	c.links = append(c.links, id)

	return nil
}

// RemoveCleanSchedule Отменяет процедуры удаления ссылок из Redis (для новой подписки)
func (c *Manager) RemoveCleanSchedule(ctx context.Context, username string) {
	ctx = log.ContextWithSpan(ctx, "RemoveCleanSchedule")
	l := c.logger.WithContext(ctx)

	l.Debug("RemoveCleanSchedule() started")
	defer l.Debug("RemoveCleanSchedule() done")

	// Итерируемся по кешу и отменяем операции
	c.mux.RLock()
	for k := 0; k < len(c.subs[username].RemId); k++ {
		c.scheduler.Remove(c.subs[username].RemId[k])
	}
	c.mux.RUnlock()

	// Чистим кэш
	c.mux.Lock()
	delete(c.subs, username)
	c.mux.Unlock()

	return
}

// CleaningUnsubscribeSchedule Планирует удаление ссылок пользователя, не соответствующих лимитам после прекращения подписки
func (c *Manager) CleanUnsubscribeSchedule(ctx context.Context, sub models.CurrentSub, username string) error {
	ctx = log.ContextWithSpan(ctx, "CleanUnsubscribeSchedule")
	l := c.logger.WithContext(ctx)

	l.Debug("CleanUnsubscribeSchedule() started")
	defer l.Debug("CleanUnsubscribeSchedule() done")

	// Считаем дату расписания
	date := formatSched(sub.Exp)

	// Создаем слайс, содержащий номера операций удаления
	sub.RemId = make([]cron.EntryID, 0)

	// Получаем кол-во ссылок пользователя по категориям
	amo, err := c.linkRepo.CountLinks(ctx, username)
	if err != nil {
		l.Errorf("Unable to count user links. Error: %s", err)
		return err
	}

	// Обьявляем счетчики обычных и кастомных ссылок
	var defaultCounter, customCounter int

	// Создаем слайсы для дальнейшей фильтрации
	defaultLinks := make([]models.LinkDataDB, 0)
	customLinks := make([]models.LinkDataDB, 0)

	// Получаем все ссылки пользователя
	links, err := c.linkRepo.GetAllLinks(ctx, username)
	if err != nil {
		l.Errorf("Unable to get all user links. Error: %s", err)
		return err
	}

	// Итерируемся по ссылкам, планируем удаление перманентных, отфильтровываем остальные
	for k := 0; k < len(links); k++ {
		if links[k].Perm {
			//
			id, err := c.scheduler.AddFunc(date, func() {
				c.linkRepo.DeleteLink(ctx, links[k].Link, username)
			})
			if err != nil {
				l.Errorf("Unable to add scheduler job. Error: %s", err)
				return err
			}

			sub.RemId = append(sub.RemId, id)

		} else {
			if links[k].Custom {
				customLinks = append(customLinks, links[k])
			} else {
				defaultLinks = append(defaultLinks, links[k])
			}
		}
	}

	// Проверяем кол-во кастомных ссылок
	if amo.Custom > Custom {
		customCounter = Custom - amo.Custom
	}

	// Итерируемся по кастомным ссылкам, планируем удаление
	for k := 0; k < customCounter; k++ {
			//
			id, err := c.scheduler.AddFunc(date, func() {
				c.linkRepo.DeleteLink(ctx, customLinks[k].Link, username)
			})
			if err != nil {
				l.Errorf("Unable to add scheduler job. Error: %s", err)
				return err
			}

			sub.RemId = append(sub.RemId, id)
	}

	// Проверяем общее кол-во ссылок
	if len(defaultLinks) + len(customLinks) - customCounter > All {
		defaultCounter = All - len(defaultLinks) - len(customLinks) + customCounter
	}

	// Итерируемся по обычным ссылкам, планируем удаление
	for k := 0; k < defaultCounter; k++ {
		//
		id, err := c.scheduler.AddFunc(date, func() {
			c.linkRepo.DeleteLink(ctx, defaultLinks[k].Link, username)
		})
		if err != nil {
			l.Errorf("Unable to add scheduler job. Error: %s", err)
			return err
		}

		sub.RemId = append(sub.RemId, id)
	}

	// Записываем данные об операциях в кэш
	c.mux.Lock()
	c.subs[username] = sub
	c.mux.Unlock()

	return nil
}
