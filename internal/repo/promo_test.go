package repo

import (
	"gophermart-loyalty/internal/app"
	"time"
)

func (suite *repoSuite) TestPromoCreate() {

	suite.Run("create", func() {
		err := suite.repo.PromoCreate(suite.ctx(), testPromo("promo100", 10, time.Now(), time.Now().Add(time.Hour)))
		suite.NoError(err)
	})

	suite.Run("retrieve", func() {
		promo, err := suite.repo.PromoGetByCode(suite.ctx(), "promo100")
		suite.NoError(err)
		suite.NotNil(promo)
		suite.Equal("10", promo.Reward.String())
	})

	suite.Run("promo_code_unique constraint", func() {
		err := suite.repo.PromoCreate(suite.ctx(), testPromo("promo100", 100, time.Now(), time.Now().Add(time.Hour*2)))
		suite.ErrorIs(err, app.ErrPromoAlreadyExists)
	})

	suite.Run("promo_reward_positive constraint", func() {
		err := suite.repo.PromoCreate(suite.ctx(), testPromo("promo200", -100, time.Now(), time.Now().Add(time.Hour*2)))
		suite.ErrorIs(err, app.ErrPromoRewardNotPositive)
	})

	suite.Run("promo_valid_period constraint", func() {
		err := suite.repo.PromoCreate(suite.ctx(), testPromo("promo300", 100, time.Now().Add(time.Hour), time.Now()))
		suite.ErrorIs(err, app.ErrPromoPeriodInvalid)
	})
}

func (suite *repoSuite) TestPromoGetByCode() {
	promo, err := suite.repo.PromoGetByCode(suite.ctx(), "non-existing")
	suite.Error(err)
	suite.ErrorIs(err, app.ErrNotFound)
	suite.Nil(promo)
}
