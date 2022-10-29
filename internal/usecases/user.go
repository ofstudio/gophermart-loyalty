package usecases

import (
	"context"
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/models"
)

var loginValidateRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._\-@ ]{2,63}$`)
var passValidateRe = regexp.MustCompile(`^.{6,512}$`)

// UserCreate - создает нового пользователя.
func (u *UseCases) UserCreate(ctx context.Context, login, password string) (*models.User, error) {
	// валидируем логин
	if !loginValidateRe.MatchString(login) {
		return nil, errs.ErrUserLoginInvalid
	}

	// валидируем пароль
	if !passValidateRe.MatchString(password) {
		return nil, errs.ErrUserPassInvalid
	}

	// Создаем хэш пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to hash password")
		return nil, errs.ErrInternal
	}

	// Создаем пользователя
	user := &models.User{
		Login:    login,
		PassHash: string(hash),
	}

	// Сохраняем пользователя
	err = u.repo.UserCreate(ctx, user)
	if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to create user")
		return nil, err
	}

	u.log.WithReqID(ctx).Debug().Uint64("user_id", user.ID).Msg("user created")
	return user, nil
}

// UserCheckLoginPass - проверяет логин и пароль пользователя.
// Возвращает пользователя, если логин и пароль верны.
func (u *UseCases) UserCheckLoginPass(ctx context.Context, login, password string) (*models.User, error) {
	// Ищем пользователя по логину
	user, err := u.repo.UserGetByLogin(ctx, login)
	if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to find user")
		return nil, errs.ErrUserLoginPassMismatch
	}
	// Сравниваем хэши паролей
	err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password))
	if err != nil {
		return nil, errs.ErrUserLoginPassMismatch
	}
	u.log.Debug().Msg("user found, password matched")
	return user, nil
}

// UserGetByID - возвращает пользователя по ID.
func (u *UseCases) UserGetByID(ctx context.Context, userID uint64) (*models.User, error) {
	user, err := u.repo.UserGetByID(ctx, userID)
	if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to find user")
		return nil, err
	}
	u.log.WithReqID(ctx).Debug().Msg("user found")
	return user, nil
}

// UserBalanceHistoryGetByID - возвращает список операций пользователя, учитывающихся в балансе.
func (u *UseCases) UserBalanceHistoryGetByID(ctx context.Context, userID uint64) ([]*models.Operation, error) {
	list, err := u.repo.UserBalanceHistoryGetByID(ctx, userID)
	if errors.Is(err, errs.ErrNotFound) {
		return nil, nil
	} else if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to get balance history")
		return nil, err
	}
	return list, nil
}
