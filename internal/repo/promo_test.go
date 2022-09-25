package repo

import (
	"gophermart-loyalty/internal/errs"
	"time"
)

func (suite *repoSuite) TestPromoCreate() {

	suite.Run("create", func() {
		err := suite.repo.PromoCreate(suite.ctx(), testPromo("promo100", 10, time.Now()))
		suite.NoError(err)
	})

	suite.Run("check", func() {
		promo, err := suite.repo.PromoGetByCode(suite.ctx(), "promo100")
		suite.NoError(err)
		suite.NotNil(promo)
		suite.Equal("10", promo.Reward.String())
	})

	suite.Run("PromoCodeUnique", func() {
		err := suite.repo.PromoCreate(suite.ctx(), testPromo("promo100", 100, time.Now()))
		suite.ErrorIs(err, errs.Duplicate)
	})

	suite.Run("PromoRewardPositive", func() {
		err := suite.repo.PromoCreate(suite.ctx(), testPromo("promo100", -100, time.Now()))
		suite.ErrorIs(err, errs.Validation)
	})
}

func (suite *repoSuite) TestPromoGetByCode() {
	promo, err := suite.repo.PromoGetByCode(suite.ctx(), "non-existing")
	suite.Error(err)
	suite.ErrorIs(err, errs.NotFound)
	suite.Nil(promo)
}
