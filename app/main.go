package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"short_url/internal/cache"
	"short_url/internal/config"
	"short_url/internal/handlers"
	"short_url/internal/handlers/middlewares"
	"short_url/internal/repositories"
	"short_url/internal/services"
	"short_url/pkg/client"
	logg "short_url/pkg/logger"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
)

func main() {
	//
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация логгера
	logger, err := logg.InitLogger(conf.Log)
	if err != nil {
		log.Fatal(err)
	}
	ctx := logg.ContextWithTrace(context.Background(), "main")
	ctx = logg.ContextWithSpan(ctx, "main")
	l := logger.WithContext(ctx)

	// Запускаем gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Инициализация клиента PostgreSQL
	db, err := client.NewPgxClient(ctx, conf.DB)
	if err != nil {
		l.Fatalf("unable connect to database. Error: %e", err)
	}

	// Инициализация клиент Redis
	redis, err := client.NewRedisClient(ctx, conf.RDB)
	if err != nil {
		l.Fatalf("unable connect to Redis. Error: %e", err)
	}
	p := redis.Pipeline()

	// Инициализация слоя repositories
	userRepo := repositories.NewPostgresqlUserRepository(&repositories.PostgresqlUserRepositoryConfig{
		Table: "user",
		DB: db,
	})
	linkRepo := repositories.NewRedisLinkRepository(&repositories.RedisLinkRepositoryConfig{
		DB: redis,
		Pipe: p,
	})
	subRepo := repositories.NewRedisSubRepository(&repositories.RedisSubRepositoryConfig{
		DB: redis,
		Pipe: p,
	})

	// Инициализация планировщика
	manager := manager.NewManager(&manager.ManagerConfig{
		LinkRepo: linkRepo,
		Scheduler: cron.New(),
		Logger: l,
	})

	// Инициализация слоя services
	tokenService := services.NewTokenService(&services.TSConfig{
		PrivateKey: conf.JWT.PrivateKey,
		PublicKey: conf.JWT.PublicKey,
		TokenExpirationSec: conf.JWT.AccessTokenExpiration,
		Logger: l,
	})
	userService := services.NewAuthService(&services.AuthServiceConfig{
		AuthRepo: userRepo,
		SubRepo: subRepo,
		Logger: l,
	})
	linkService := services.NewLinkService(&services.LinkServiceConfig{
		LinkRepo: linkRepo,
		Manager: manager,
	})
	qiwiService := services.NewQiwiService(&services.QiwiServiceConfig{
		Key: conf.App.SecretKey,
		SubRepo: subRepo,
		AuthRepo: userRepo,
		Manager: manager,
		Prices: *conf.Price,
		Logger: l,
	})

	// Регистрация middleware
	middleware := middlewares.NewMiddlewares(l, tokenService)

	// Регистрация счетчика Prometheus
	prometheus.MustRegister(middleware.Counter)

	// Инициализация слоя handlers
	handlers.RegisterAuthHandler(&handlers.AuthHandlerConfig{
		Router: router,
		AuthService: userService,
		TokenService: tokenService,
		Middleware: middleware,
		Logger: l,
	})
	handlers.RegisterLinkHandler(&handlers.LinkHandlerConfig{
		Router: router,
		LinkService: linkService,
		Middleware: middleware,
		Logger: l,
	})
	handlers.RegisterPayHandler(&handlers.PayHandlerConfig{
		Router: router,
		QiwiService: qiwiService,
		Key: conf.App.SecretKey,
		Prices: *conf.Price,
		Middleware: middleware,
		Logger: l,
	})

	// Запуск фонового процесса планировщика
	schedChan := manager.SchedChecker(ctx)

	// Запуск фонового процесса системы проверки платежей Qiwi
	qiwiChan := qiwiService.QiwiCheckCycle(ctx)

	// Инициализация основного сервера
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%s", conf.HTTP.Host, conf.HTTP.Port),
		Handler: router,
	}

	// Инициализация сервера метрик
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", conf.HTTP.Host, conf.HTTP.MetricPort),
		Handler: metricsMux,
	}

	// Запуск серверов
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to initialize server: %v\n", err)
		}
	}()

	go func() {
		if err := metricServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to initialize metrics server: %v\n", err)
		}
	}()

	l.Infof("server listening on port %v", conf.HTTP.Port)
	l.Infof("prometheus listening on port %v", conf.HTTP.MetricPort)

	// Graceful shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	l.Info("shutting down qiwi-pay system...")
	qiwiChan <- struct{}{}

	l.Info("shutting down scheduler system...")
	schedChan <- struct{}{}

	l.Info("shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v\n", err)
	}

	l.Info("shutting down metrics server...")
	if err := metricServer.Shutdown(ctx); err != nil {
		log.Fatalf("metrics server forced to shutdown: %v\n", err)
	}

	<-ctx.Done()
	l.Info("successfully")
}
