package usecases

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/models"
)

func (suite *useCasesSuite) TestUserCreate() {
	password := "Qwerty123456!"

	suite.Run("success", func() {
		suite.repo.On("UserCreate",
			mock.Anything,
			mock.MatchedBy(func(user *models.User) bool {
				err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password))
				return err == nil
			})).
			Return(nil).
			Run(func(args mock.Arguments) {
				user := args.Get(1).(*models.User)
				user.ID = 10
			}).
			Once()
		user, err := suite.useCases.UserCreate(suite.ctx(), "oleg", password)
		suite.NoError(err)
		suite.Equal("oleg", user.Login)
		suite.Equal(uint64(10), user.ID)
	})

	suite.Run("user already exists", func() {
		suite.repo.On("UserCreate", mock.Anything, mock.Anything).
			Return(errs.ErrUserAlreadyExists).Once()
		user, err := suite.useCases.UserCreate(suite.ctx(), "oleg", password)
		suite.ErrorIs(err, errs.ErrUserAlreadyExists)
		suite.Nil(user)
	})

	suite.Run("password too short", func() {
		user, err := suite.useCases.UserCreate(suite.ctx(), "oleg", "12345")
		suite.ErrorIs(err, errs.ErrUserPassInvalid)
		suite.Nil(user)
	})

	suite.Run("password too long", func() {
		user, err := suite.useCases.UserCreate(suite.ctx(), "oleg", strings.Repeat("1", 513))
		suite.ErrorIs(err, errs.ErrUserPassInvalid)
		suite.Nil(user)
	})

	suite.Run("login too short", func() {
		user, err := suite.useCases.UserCreate(suite.ctx(), "of", password)
		suite.ErrorIs(err, errs.ErrUserLoginInvalid)
		suite.Nil(user)
	})

	suite.Run("login too long", func() {
		user, err := suite.useCases.UserCreate(suite.ctx(), strings.Repeat("o", 65), password)
		suite.ErrorIs(err, errs.ErrUserLoginInvalid)
		suite.Nil(user)
	})

	suite.Run("login contains invalid characters", func() {
		user, err := suite.useCases.UserCreate(suite.ctx(), "o!leg", password)
		suite.ErrorIs(err, errs.ErrUserLoginInvalid)
		suite.Nil(user)
	})

	suite.Run("login starts with invalid characters", func() {
		user, err := suite.useCases.UserCreate(suite.ctx(), "-oleg", password)
		suite.ErrorIs(err, errs.ErrUserLoginInvalid)
		suite.Nil(user)
	})
}

func (suite *useCasesSuite) TestUserCheckLoginPass() {
	password := "Qwerty123456!"
	passhash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	suite.Run("success", func() {
		suite.repo.On("UserGetByLogin", mock.Anything, "oleg").
			Return(&models.User{
				ID:       1,
				Login:    "oleg",
				PassHash: string(passhash),
			}, nil).Once()
		user, err := suite.useCases.UserCheckLoginPass(suite.ctx(), "oleg", password)
		suite.NoError(err)
		suite.Equal("oleg", user.Login)
	})

	suite.Run("user not found", func() {
		suite.repo.On("UserGetByLogin", mock.Anything, "oleg").
			Return(nil, errs.ErrNotFound).Once()
		user, err := suite.useCases.UserCheckLoginPass(suite.ctx(), "oleg", password)
		suite.ErrorIs(err, errs.ErrUserLoginPassMismatch)
		suite.Nil(user)
	})

	suite.Run("password mismatch", func() {
		suite.repo.On("UserGetByLogin", mock.Anything, "oleg").
			Return(&models.User{
				ID:       1,
				Login:    "oleg",
				PassHash: string(passhash),
			}, nil).Once()
		user, err := suite.useCases.UserCheckLoginPass(suite.ctx(), "oleg", "wrong password")
		suite.ErrorIs(err, errs.ErrUserLoginPassMismatch)
		suite.Nil(user)
	})
}

func (suite *useCasesSuite) TestUserGetByID() {
	suite.Run("success", func() {
		suite.repo.On("UserGetByID", mock.Anything, mock.Anything).
			Return(&models.User{
				ID:    1,
				Login: "oleg",
			}, nil).Once()
		user, err := suite.useCases.UserGetByID(suite.ctx(), 1)
		suite.NoError(err)
		suite.Equal("oleg", user.Login)
	})

	suite.Run("user not found", func() {
		suite.repo.On("UserGetByID", mock.Anything, mock.Anything).
			Return(nil, errs.ErrNotFound).Once()
		user, err := suite.useCases.UserGetByID(suite.ctx(), 1)
		suite.ErrorIs(err, errs.ErrNotFound)
		suite.Nil(user)
	})
}

func (suite *useCasesSuite) TestUserBalanceHistoryGetByID() {
	suite.Run("success", func() {
		ops := []*models.Operation{
			{
				ID:          1,
				UserID:      1,
				Type:        models.OrderAccrual,
				Status:      models.StatusNew,
				Amount:      decimal.NewFromInt(100),
				Description: "Начисление баллов за заказ 1",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		suite.repo.On("UserBalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(ops, nil).Once()
		history, err := suite.useCases.UserBalanceHistoryGetByID(suite.ctx(), uint64(1))
		suite.NoError(err)
		suite.Equal(ops, history)
	})

	suite.Run("no operations", func() {
		suite.repo.On("UserBalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(nil, errs.ErrNotFound).Once()
		history, err := suite.useCases.UserBalanceHistoryGetByID(suite.ctx(), uint64(1))
		suite.NoError(err)
		suite.Nil(history)
	})

	suite.Run("internal error", func() {
		suite.repo.On("UserBalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(nil, errs.ErrInternal).Once()
		history, err := suite.useCases.UserBalanceHistoryGetByID(suite.ctx(), uint64(1))
		suite.ErrorIs(err, errs.ErrInternal)
		suite.Nil(history)
	})

}
