package usecases

import (
	"context"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
	"regexp"
)

var LoginValidateRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._\-@ ]{2,63}$`)
var PassValidateRe = regexp.MustCompile(`^.{6,512}$`)

func (u *UseCases) UserCreate(ctx context.Context, login, password string) (*models.User, error) {
	log := u.log.WithReqID(ctx)

	// проверяем логин
	if !LoginValidateRe.MatchString(login) {
		return nil, app.ErrUserLoginInvalid
	}

	// проверяем пароль
	if !PassValidateRe.MatchString(password) {
		return nil, app.ErrUserPassInvalid
	}

	// Создаем хэш пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash password")
		return nil, app.ErrInternal
	}

	// Создаем пользователя
	user := &models.User{
		Login:    login,
		PassHash: string(hash),
	}

	// Сохраняем пользователя
	err = u.repo.UserCreate(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("failed to create user")
		return nil, err
	}

	log.Debug().Uint64("user_id", user.ID).Msg("user created")
	return user, nil
}

func (u *UseCases) UserCheckLoginPass(ctx context.Context, login, password string) (*models.User, error) {

	// Ищем пользователя по логину
	user, err := u.repo.UserGetByLogin(ctx, login)
	if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to find user")
		return nil, app.ErrUnauthorized
	}

	// Сравниваем хэши паролей
	err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password))
	if err != nil {
		return nil, app.ErrUnauthorized
	}
	log.Debug().Msg("user found, password matched")
	return user, nil
}

func (u *UseCases) UserGetByID(ctx context.Context, userID uint64) (*models.User, error) {
	log := u.log.WithReqID(ctx).With().Uint64("user_id", userID).Logger()
	user, err := u.repo.UserGetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to find user")
		return nil, err
	}
	log.Debug().Msg("user found")
	return user, nil
}
