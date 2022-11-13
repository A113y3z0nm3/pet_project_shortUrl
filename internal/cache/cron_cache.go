package cache

import (
	"time"
	"sync"
	"context"
	"short_url/internal/models"
	cron "github.com/robfig/cron/v3"
)

const (
	All    			= 50				// Лимит кол-ва всех ссылок для обычного пользователя
	//SubAll 			= 100				// Лимит кол-ва всех ссылок для подписчика

	Custom    		= 15				// Лимит кол-ва кастомных ссылок для обычного пользователя
	//SubCustom 		= 30				// Лимит кол-ва кастомных ссылок для подписчика

	SubPerm			= 10				// Лимит кол-ва ссылок с безграничным сроком действия для подписчика
)

// linkRepository Интерфейс к репозиторию управления ссылками
type linkRepository interface {
	DeleteExpLink(ctx context.Context, link, username string) error
	DeleteLink(ctx context.Context, link, username string) error
	CountLinks(ctx context.Context, username string) (models.LinksAmount, error)
	GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDB, error)
}

// CronCacheConfig 
type CronCacheConfig struct {
	LinkRepo		linkRepository
	Scheduler		*cron.Cron
}

// CronCache 
type CronCache struct {
	linkRepo		linkRepository
	scheduler		*cron.Cron
	subs			map[string]models.CurrentSub
	mux				sync.RWMutex
}

// NewCronCache 
func NewCronCache(conf *CronCacheConfig) *CronCache {
	return &CronCache{
		linkRepo: conf.LinkRepo,
		scheduler: conf.Scheduler,
		subs: make(map[string]models.CurrentSub),
	}
}

// CleaningSchedule 
func (c *CronCache) CleaningExpLinkSchedule(ctx context.Context, link, username string, exp time.Duration) {
	//
	c.scheduler.AddFunc("", func(){
		c.linkRepo.DeleteExpLink(ctx, link, username)
	})

	return
}

// RemoveCleanSchedule 
func (c *CronCache) RemoveCleanSchedule(username string) {
	for _, id := range c.subs[username].RemId {
		c.scheduler.Remove(id)
	}
}

// CleaningUnsubscribeSchedule 
func (c *CronCache) CleanUnsubscribeSchedule(ctx context.Context, sub models.CurrentSub, username string) error {
	//
	sub.RemId = make([]cron.EntryID, 0)

	//
	amo, err := c.linkRepo.CountLinks(ctx, username)
	if err != nil {
		// log
		return err
	}

	//
	var counter, customCounter int

	//
	defaultLinks := make([]models.LinkDataDB, 0)
	customLinks := make([]models.LinkDataDB, 0)

	//
	links, err := c.linkRepo.GetAllLinks(ctx, username)
	if err != nil {
		// log
		return err
	}

	//
	for k := 0; k < len(links); k++ {
		if links[k].Perm {
			//
			id, err := c.scheduler.AddFunc("////////", func() {
				c.linkRepo.DeleteLink(ctx, links[k].Link, username)
			})
			if err != nil {
				// log
				return err
			}

			c.mux.Lock()
			sub.RemId = append(c.subs[username].RemId, id)
			c.mux.Unlock()

		} else {
			if links[k].Custom {
				customLinks = append(customLinks, links[k])
			} else {
				defaultLinks = append(defaultLinks, links[k])
			}
		}
	}

	//
	if amo.Custom > Custom {
		customCounter = Custom - amo.Custom
	}

	//
	for k := 0; k < customCounter; k++ {
			//
			id, err := c.scheduler.AddFunc("////////", func() {
				c.linkRepo.DeleteLink(ctx, customLinks[k].Link, username)
			})
			if err != nil {
				// log
				return err
			}

			c.mux.Lock()
			sub.RemId = append(c.subs[username].RemId, id)
			c.mux.Unlock()
	}

	//
	if len(defaultLinks) + len(customLinks) - customCounter > All {
		counter = All - len(defaultLinks) - len(customLinks) + customCounter
	}

	//
	for k := 0; k < counter; k++ {
		//
		id, err := c.scheduler.AddFunc("////////", func() {
			c.linkRepo.DeleteLink(ctx, defaultLinks[k].Link, username)
		})
		if err != nil {
			// log
			return err
		}

		c.mux.Lock()
		sub.RemId = append(c.subs[username].RemId, id)
		c.mux.Unlock()
	}

	return nil
}
