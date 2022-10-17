// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"
	models "gophermart-loyalty/internal/models"

	mock "github.com/stretchr/testify/mock"

	repo "gophermart-loyalty/internal/repo"
)

// Repo is an autogenerated mock type for the Repo type
type Repo struct {
	mock.Mock
}

// BalanceHistoryGetByID provides a mock function with given fields: ctx, userID
func (_m *Repo) BalanceHistoryGetByID(ctx context.Context, userID uint64) ([]*models.Operation, error) {
	ret := _m.Called(ctx, userID)

	var r0 []*models.Operation
	if rf, ok := ret.Get(0).(func(context.Context, uint64) []*models.Operation); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Operation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint64) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Close provides a mock function with given fields:
func (_m *Repo) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OperationCreate provides a mock function with given fields: ctx, op
func (_m *Repo) OperationCreate(ctx context.Context, op *models.Operation) error {
	ret := _m.Called(ctx, op)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Operation) error); ok {
		r0 = rf(ctx, op)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OperationGetByType provides a mock function with given fields: ctx, userID, t
func (_m *Repo) OperationGetByType(ctx context.Context, userID uint64, t models.OperationType) ([]*models.Operation, error) {
	ret := _m.Called(ctx, userID, t)

	var r0 []*models.Operation
	if rf, ok := ret.Get(0).(func(context.Context, uint64, models.OperationType) []*models.Operation); ok {
		r0 = rf(ctx, userID, t)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Operation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint64, models.OperationType) error); ok {
		r1 = rf(ctx, userID, t)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OperationUpdateFurther provides a mock function with given fields: ctx, opType, updateFunc
func (_m *Repo) OperationUpdateFurther(ctx context.Context, opType models.OperationType, updateFunc repo.UpdateFunc) (*models.Operation, error) {
	ret := _m.Called(ctx, opType, updateFunc)

	var r0 *models.Operation
	if rf, ok := ret.Get(0).(func(context.Context, models.OperationType, repo.UpdateFunc) *models.Operation); ok {
		r0 = rf(ctx, opType, updateFunc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Operation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.OperationType, repo.UpdateFunc) error); ok {
		r1 = rf(ctx, opType, updateFunc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PromoCreate provides a mock function with given fields: ctx, p
func (_m *Repo) PromoCreate(ctx context.Context, p *models.Promo) error {
	ret := _m.Called(ctx, p)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Promo) error); ok {
		r0 = rf(ctx, p)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PromoGetByCode provides a mock function with given fields: ctx, code
func (_m *Repo) PromoGetByCode(ctx context.Context, code string) (*models.Promo, error) {
	ret := _m.Called(ctx, code)

	var r0 *models.Promo
	if rf, ok := ret.Get(0).(func(context.Context, string) *models.Promo); ok {
		r0 = rf(ctx, code)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Promo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, code)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserCreate provides a mock function with given fields: ctx, u
func (_m *Repo) UserCreate(ctx context.Context, u *models.User) error {
	ret := _m.Called(ctx, u)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.User) error); ok {
		r0 = rf(ctx, u)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UserGetByID provides a mock function with given fields: ctx, userID
func (_m *Repo) UserGetByID(ctx context.Context, userID uint64) (*models.User, error) {
	ret := _m.Called(ctx, userID)

	var r0 *models.User
	if rf, ok := ret.Get(0).(func(context.Context, uint64) *models.User); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint64) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserGetByLogin provides a mock function with given fields: ctx, login
func (_m *Repo) UserGetByLogin(ctx context.Context, login string) (*models.User, error) {
	ret := _m.Called(ctx, login)

	var r0 *models.User
	if rf, ok := ret.Get(0).(func(context.Context, string) *models.User); ok {
		r0 = rf(ctx, login)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewRepo interface {
	mock.TestingT
	Cleanup(func())
}

// NewRepo creates a new instance of Repo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRepo(t mockConstructorTestingTNewRepo) *Repo {
	mock := &Repo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
