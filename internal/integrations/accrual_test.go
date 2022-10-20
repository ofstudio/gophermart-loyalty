package integrations

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/mocks"
	"gophermart-loyalty/internal/models"
	"gophermart-loyalty/internal/repo"
	"gophermart-loyalty/internal/usecases"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testPollInterval = 500 * time.Millisecond
	testTimeout      = 1 * time.Second
)

func TestAccrualSuite(t *testing.T) {
	suite.Run(t, new(accrualSuite))
}

/*
- [x] Start / Stop
- [x] Successful request
- [x] Failed request
- [x] Timeout
- [x] Too many requests
- [x] No operations to update
- [x] Update failed
*/

func (suite *accrualSuite) TestStartStop() {
	ctx, cancel := context.WithCancel(suite.ctx())
	defer cancel()
	suite.accrual.Start(ctx)
	suite.Equal(AccrualRunning, suite.accrual.Status())
	time.Sleep(testPollInterval / 2)
	cancel()
	time.Sleep(testPollInterval)
	suite.Equal(AccrualStopped, suite.accrual.Status())
}

func (suite *accrualSuite) TestSuccessfulRequest() {
	suite.testHandler = suite.handlers["success"]
	suite.mockCalls["success"]().Once()
	ctx, cancel := context.WithCancel(suite.ctx())
	defer cancel()
	suite.accrual.Start(ctx)
	time.Sleep(testPollInterval + 100*time.Millisecond)
}

func (suite *accrualSuite) TestFailedRequest() {
	suite.testHandler = suite.handlers["failed"]
	suite.mockCalls["success"]().Once()
	err := suite.accrual.updateFurther(suite.ctx())
	suite.ErrorIs(err, app.ErrIntegrationRequestFailed)
}

func (suite *accrualSuite) TestTimeout() {
	suite.testHandler = suite.handlers["timeout"]
	suite.mockCalls["success"]().Once()
	err := suite.accrual.updateFurther(suite.ctx())
	suite.ErrorIs(err, app.ErrIntegrationRequestFailed)
}

func (suite *accrualSuite) TestTooManyRequests() {
	suite.testHandler = suite.handlers["too_many_requests"]
	suite.mockCalls["success"]().Once()
	ctx, cancel := context.WithCancel(suite.ctx())
	defer cancel()
	suite.accrual.Start(ctx)
	time.Sleep(2 * time.Second)
	suite.Equal(1*time.Second, suite.accrual.pollInterval)
	suite.Equal(0*time.Second, suite.accrual.retryAfter)
}

func (suite *accrualSuite) TestNoOperationsToUpdate() {
	suite.mockCalls["no_operations_to_update"]().Once()
	err := suite.accrual.updateFurther(suite.ctx())
	suite.NoError(err)
}

func (suite *accrualSuite) TestUpdateFailed() {
	suite.mockCalls["failed"]().Once()
	err := suite.accrual.updateFurther(suite.ctx())
	suite.ErrorIs(err, app.ErrInternal)
}

type accrualSuite struct {
	suite.Suite
	accrual     *Accrual
	useCases    *usecases.UseCases
	repo        *mocks.Repo
	log         logger.Log
	testServer  *httptest.Server
	testHandler http.HandlerFunc
	handlers    map[string]http.HandlerFunc
	mockCalls   map[string]func() *mock.Call
}

func (suite *accrualSuite) SetupSuite() {
	suite.log = logger.NewLogger(zerolog.DebugLevel)

	suite.handlers = map[string]http.HandlerFunc{
		"success": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order": "2377225624", "status": "PROCESSING"}`))
		},
		"failed": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "bad request"}`))
		},
		"timeout": func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Second)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order": "2377225624", "status": "PROCESSING"}`))
		},
		"too_many_requests": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`No more than 60 requests per minute allowed`))
		},
	}

	suite.mockCalls = map[string]func() *mock.Call{
		"success": func() *mock.Call {
			c := suite.repo.
				On("OperationUpdateFurther", mock.Anything, models.OrderAccrual, mock.Anything).
				Return(&models.Operation{ID: 1}, nil)
			c.RunFn = func(args mock.Arguments) {
				ctx := args.Get(0).(context.Context)
				updateFunc := args.Get(2).(repo.UpdateFunc)
				err := updateFunc(ctx, &models.Operation{ID: 1, OrderNumber: strPtr("2377225624")})
				if err != nil {
					c.ReturnArguments = mock.Arguments{nil, err}
					return
				}
				c.ReturnArguments = mock.Arguments{&models.Operation{ID: 1, OrderNumber: strPtr("2377225624")}, nil}
			}
			return c
		},
		"no_operations_to_update": func() *mock.Call {
			return suite.repo.
				On("OperationUpdateFurther", mock.Anything, models.OrderAccrual, mock.Anything).
				Return(nil, app.ErrNotFound)
		},
		"failed": func() *mock.Call {
			return suite.repo.
				On("OperationUpdateFurther", mock.Anything, models.OrderAccrual, mock.Anything).
				Return(nil, app.ErrInternal)
		},
	}
}

func (suite *accrualSuite) SetupTest() {
	suite.testHandler = nil
	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		suite.testHandler(w, r)
	})
	suite.testServer = httptest.NewServer(mux)

	suite.repo = mocks.NewRepo(suite.T())
	suite.useCases = usecases.NewUseCases(suite.repo, suite.log)
	cfg := &config.IntegrationAccrual{
		Address:      suite.testServer.URL,
		PollInterval: testPollInterval,
		Timeout:      testTimeout,
	}
	suite.accrual = NewAccrual(cfg, suite.useCases, suite.log)
}

func (suite *accrualSuite) TearDownTest() {
	suite.testServer.Close()
}

func (suite *accrualSuite) ctx() context.Context {
	return context.WithValue(context.Background(), middleware.RequestIDKey, suite.T().Name())
}

func strPtr(s string) *string {
	return &s
}
