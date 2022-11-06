package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"short_url/internal/models"
	"short_url/internal/security"
	myLog "short_url/pkg/logger"
)

// authRepository Интерфейс к репозиторию хранения данных пользователей
type authRepository interface {
	FindByUsername(ctx context.Context, name string) (models.UserDB, error)
	CreateUser(ctx context.Context, user models.UserDB) error
}

// AuthServiceConfig Конфигурация к AuthService
type AuthServiceConfig struct {
	AuthRepo authRepository
	Logger   *myLog.Log
}

// AuthService Управляет регистрацией и аутентификацией пользователей
type AuthService struct {
	authRepo authRepository
	logger   *myLog.Log
}

// Конструктор для AuthService
func NewAuthService(c *AuthServiceConfig) *AuthService {
	return &AuthService{
		authRepo: c.AuthRepo,
		logger:   c.Logger,
	}
}

// SignInUserByName вызывает методы других слоев, которые позволят войти пользователю по его имени
func (s *AuthService) SignInUserByName(ctx context.Context, dto models.SignInUserDTO) (models.SignInUserDTO, error) {
	ctx = myLog.ContextWithSpan(ctx, "SignInUserByName")
	l := s.logger.WithContext(ctx)

	l.Debug("SignInUserByName() started")
	defer l.Debug("SignInUserByName() done")

	// Ищем зарегистрированного пользователя
	u, err := s.authRepo.FindByUsername(ctx, dto.Username)
	if err != nil {
		l.Errorf("unable to find user. Err: %e", err)

		return dto, err
	}

	// Сравниваем пароли пользователя
	ok, err := security.ComparePasswords(u.Password, dto.Password)
	if err != nil {
		l.Errorf("unable to compare password. Err: %e", err)

		return dto, err
	}
	if !ok {
		return dto, fmt.Errorf("invalid password")
	}

	// Мапим данные из db в dto структуру
	dto.FirstName	= u.FirstName
	dto.LastName	= u.LastName
	dto.Subscribe	= dto.Subscribe.ChoiceSubscribe(u.Subscribe)

	return dto, nil
}

// SignUpUser вызывает методы для создания пользователя и хеширования пароля
func (s *AuthService) SignUpUser(ctx context.Context, dto models.SignUpUserDTO) error {
	ctx = myLog.ContextWithSpan(ctx, "SignUpUser")
	l := s.logger.WithContext(ctx)

	l.Debug("SignUpUser() started")
	defer l.Debug("SignUpUser() done")

	// Проверяем, зарегистрирован ли пользователь
	_, err := s.authRepo.FindByUsername(ctx, dto.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			l.Info("user not found")
		} else {
			l.Errorf("unable to find user. Err: %e", err)

			return err
		}
	} else {
		return fmt.Errorf("user exist")
	}

	// Хешируем пароль
	hashPassword, err := security.HashPassword(dto.Password)
	if err != nil {
		l.Errorf("unable to hash password. Err: %e", err)

		return err
	}

	// Создаем пользователя в базе данных
	err = s.authRepo.CreateUser(ctx, models.UserDB{
		Username:	dto.Username,
		FirstName:	dto.FirstName,
		LastName:	dto.LastName,
		Password:	hashPassword,
		Subscribe:	dto.Subscribe.ChoiceString(),
	})
	if err != nil {
		l.Errorf("unable to create user. Err: %e", err)

		return err
	}

	return nil
}