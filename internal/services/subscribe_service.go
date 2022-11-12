package services

// import (
// 	"fmt"
// 	"short_url/internal/models"
// 	"time"

// 	cron "github.com/robfig/cron/v3"
// )

// //
// type SubServiceConfig struct {
// 	UserRepo	userRepository
// 	SubRepo		subRepository
// }

// //
// type SubService struct {
// 	userRepo	userRepository
// 	subRepo		subRepository
// 	scheduler	*cron.Cron
// 	schedCh		chan cron.EntryID
// }

// //
// func NewSubService(c *SubServiceConfig) *SubService {
// 	localTime, _ := time.LoadLocation("Local")

// 	return &SubService{
// 		userRepo:	c.UserRepo,
// 		subRepo:	c.SubRepo,
// 		scheduler:	cron.New(cron.WithLocation(localTime)),
// 		schedCh:	make(chan cron.EntryID),
// 	}
// }

// //
// func (s *SubService) Subscribe() {

// }

// //
// func (s *SubService) CheckSubscribe(user models.UserDB) {
// 	//
// 	id := <-s.schedCh


	
// }

// //
// func (s *SubService) SubscribeSched(user models.UserDB) error {
// 	//
// 	date := time.Now()
// 	day := date.Day()
// 	min := date.Minute()
// 	hour := date.Hour()
	
// 	//
// 	id, err := s.scheduler.AddFunc(fmt.Sprint("%s %s %s * *", min, hour, day), func(){ s.DeleteSubscribe(user) })
// 	if err != nil {
// 		return err
// 	}

// 	//
// 	s.schedCh <- id
// }
