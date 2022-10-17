package config

import (
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(configSuite))
}

type configSuite struct {
	suite.Suite
}

func (suite *configSuite) TestConfig() {
	suite.Run("success from cli", func() {
		cfg, err := Compose(Default)
		suite.NoError(err)
		suite.NotNil(cfg)
		args := []string{
			"-a", "0.0.0.0:3000",
			"-d", "postgres://autotest:autotest@localhost:5432/autotest",
			"-r", "http://localhost:5000",
			"-t", "10h",
		}

		cfg, err = fromCLI(cfg, args...)
		suite.NoError(err)
		suite.NotNil(cfg)
		suite.Equal("0.0.0.0:3000", cfg.RunAddress)
		suite.Equal("postgres://autotest:autotest@localhost:5432/autotest", cfg.DB.URI)
		suite.Equal("http://localhost:5000", cfg.IntegrationAccrual.Address)
		suite.Equal(10*time.Hour, cfg.Auth.TTL)
	})

	suite.Run("success from env", func() {
		os.Clearenv()
		cfg, err := Compose(Default)
		suite.NoError(err)
		suite.NotNil(cfg)

		_ = os.Setenv("RUN_ADDRESS", "0.0.0.0:9000")
		_ = os.Setenv("DATABASE_URI", "postgres://autotest2:autotest2@localhost:5432/autotest2")
		_ = os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:5001")
		_ = os.Setenv("AUTH_TTL", "2h")
		_ = os.Setenv("AUTH_SECRET", "<this is secret>")

		cfg, err = FromEnv(cfg)
		suite.NoError(err)
		suite.NotNil(cfg)
		suite.Equal("0.0.0.0:9000", cfg.RunAddress)
		suite.Equal("postgres://autotest2:autotest2@localhost:5432/autotest2", cfg.DB.URI)
		suite.Equal("http://localhost:5001", cfg.IntegrationAccrual.Address)
		suite.Equal(2*time.Hour, cfg.Auth.TTL)
	})
}
