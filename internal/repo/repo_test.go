package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/models"
)

// Если тестовая БД не будет запущена, то тесты будут пропущены
const autotestDSN = "postgres://autotest:autotest@localhost:5432/autotest"

func TestPGXRepoSuite(t *testing.T) {
	suite.Run(t, new(pgxRepoSuite))
}

type pgxRepoSuite struct {
	suite.Suite
	repo  *PGXRepo
	log   logger.Log
	stmts map[string]*sql.Stmt
}

func (suite *pgxRepoSuite) SetupSuite() {
	suite.log = logger.NewLogger(zerolog.DebugLevel)
	// Если тестовая БД не запущена - пропускаем тест
	if err := suite.isDBAvailable(autotestDSN); err != nil {
		suite.T().Skipf("skipping suite: database is not available: %v", err)
		return
	}
}

func (suite *pgxRepoSuite) SetupTest() {
	//var err error
	// Очищаем тестовую БД
	suite.NoError(suite.clearDB(autotestDSN))

	// Создаем репозиторий
	var err error
	suite.repo, err = NewPGXRepo(&config.DB{URI: autotestDSN, RequiredVersion: 1}, suite.log)
	suite.NoError(err)

	// Создаем пользователей
	suite.NoError(suite.repo.UserCreate(suite.ctx(), &models.User{Login: "user1", PassHash: "hash1"}))
	suite.NoError(suite.repo.UserCreate(suite.ctx(), &models.User{Login: "user2", PassHash: "hash2"}))
	suite.NoError(suite.repo.UserCreate(suite.ctx(), &models.User{Login: "user3", PassHash: "hash3"}))

	// Создаем промо-кампании
	suite.NoError(suite.repo.PromoCreate(suite.ctx(), &models.Promo{
		Code:        "TEST-PROMO",
		Description: "Test promo",
		Reward:      decimal.NewFromInt(5),
		NotBefore:   time.Now().Add(-time.Hour * 24),
		NotAfter:    time.Now().Add(time.Hour * 24 * 7),
	}))
}

func (suite *pgxRepoSuite) TearDownTest() {
	// Закрываем соединение
	suite.repo.Close()
}

func (suite *pgxRepoSuite) ctx() context.Context {
	return context.WithValue(context.Background(), middleware.RequestIDKey, suite.T().Name())
}

func (suite *pgxRepoSuite) isDBAvailable(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	//goland:noinspection ALL
	defer db.Close()
	return db.Ping()
}

func (suite *pgxRepoSuite) clearDB(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	//goland:noinspection ALL
	defer db.Close()
	if err = goose.SetDialect("pgx"); err != nil {
		return err
	}
	ver, err := goose.GetDBVersion(db)
	if err != nil {
		return err
	}
	if ver == 0 {
		return nil
	}
	return goose.Down(db, ".")
}

func testOA(u uint64, n string, a int, s models.OperationStatus) *models.Operation {
	return &models.Operation{
		UserID:      u,
		Type:        models.OrderAccrual,
		Status:      s,
		Amount:      decimal.NewFromInt(int64(a)),
		Description: "test",
		OrderNumber: &n,
	}
}
func testOW(u uint64, n string, a int, s models.OperationStatus) *models.Operation {
	return &models.Operation{
		UserID:      u,
		Type:        models.OrderWithdrawal,
		Status:      s,
		Amount:      decimal.NewFromInt(int64(a)),
		Description: "test",
		OrderNumber: &n,
	}
}

func testPA(u uint64, p uint64, a int, s models.OperationStatus) *models.Operation {
	return &models.Operation{
		UserID:      u,
		Type:        models.PromoAccrual,
		Status:      s,
		Amount:      decimal.NewFromInt(int64(a)),
		Description: "test",
		PromoID:     &p,
	}
}

func testPromo(code string, reward int, notBefore, notAfter time.Time) *models.Promo {
	return &models.Promo{
		Code:        code,
		Description: "test",
		Reward:      decimal.NewFromInt(int64(reward)),
		NotBefore:   notBefore,
		NotAfter:    notAfter,
	}
}
