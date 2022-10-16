package repo

import (
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
)

func (suite *pgxRepoSuite) TestUserCreate() {
	suite.NotNil(suite.repo)
	user := &models.User{Login: "user100", PassHash: "hash100"}
	suite.NoError(suite.repo.UserCreate(suite.ctx(), user))
	suite.NotZero(user.ID)
	user, err := suite.repo.UserGetByLogin(suite.ctx(), "user100")
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("hash100", user.PassHash)
	err = suite.repo.UserCreate(suite.ctx(), &models.User{Login: "user100", PassHash: "hash100"})
	suite.ErrorIs(err, app.ErrUserAlreadyExists)
}

func (suite *pgxRepoSuite) TestUserGetByLogin() {
	user, err := suite.repo.UserGetByLogin(suite.ctx(), "user1")
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("user1", user.Login)
	user, err = suite.repo.UserGetByLogin(suite.ctx(), "user1000")
	suite.Error(err)
	suite.ErrorIs(err, app.ErrNotFound)
}

func (suite *pgxRepoSuite) TestUserGetByID() {
	user, err := suite.repo.UserGetByID(suite.ctx(), 1)
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("user1", user.Login)
	user, err = suite.repo.UserGetByID(suite.ctx(), 1000)
	suite.Error(err)
	suite.ErrorIs(err, app.ErrNotFound)
}
