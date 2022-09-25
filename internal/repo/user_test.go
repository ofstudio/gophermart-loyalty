package repo

import (
	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/models"
)

func (suite *repoSuite) TestUserCreate() {
	suite.NotNil(suite.repo)
	user := &models.User{Login: "user100", PassHash: "hash100"}
	suite.NoError(suite.repo.UserCreate(suite.ctx(), user))
	suite.NotZero(user.ID)
	user, err := suite.repo.UserGetByLogin(suite.ctx(), "user100")
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("hash100", user.PassHash)
	err = suite.repo.UserCreate(suite.ctx(), &models.User{Login: "user100", PassHash: "hash100"})
	suite.ErrorIs(err, errs.Duplicate)
}

func (suite *repoSuite) TestUserGetByLogin() {
	user, err := suite.repo.UserGetByLogin(suite.ctx(), "user1")
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("user1", user.Login)
	user, err = suite.repo.UserGetByLogin(suite.ctx(), "user1000")
	suite.Error(err)
	suite.ErrorIs(err, errs.NotFound)
}

func (suite *repoSuite) TestUserGetByID() {
	user, err := suite.repo.UserGetByID(suite.ctx(), 1)
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("user1", user.Login)
	user, err = suite.repo.UserGetByID(suite.ctx(), 1000)
	suite.Error(err)
	suite.ErrorIs(err, errs.NotFound)
}
