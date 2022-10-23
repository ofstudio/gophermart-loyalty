// Package errs - бизнес-ошибки.
package errs

import (
	"errors"
)

// Общие ошибки приложения 1000-1099
var (
	// ErrInternal - внутренняя ошибка сервера
	ErrInternal = NewError(1000, 500, "Internal error")
	// ErrNotFound - не найдено
	ErrNotFound = NewError(1001, 404, "Not found")
	// ErrUnauthorized - пользователь не авторизован
	ErrUnauthorized = NewError(1002, 401, "Unauthorized")
	// ErrBadRequest - неверный запрос
	ErrBadRequest = NewError(1003, 400, "Bad request")
)

// Ошибки пользователя 1100-1199
var (
	// ErrUserAlreadyExists - логин должен быть уникальным
	ErrUserAlreadyExists = NewError(1100, 409, "User already exists")

	// ErrUserLoginInvalid - недопустимый логин
	ErrUserLoginInvalid = NewError(1101, 400, "Invalid login")

	// ErrUserPassInvalid - недопустимый пароль
	ErrUserPassInvalid = NewError(1102, 400, "Invalid password")

	// ErrUserLoginPassMismatch - неверный логин или пароль
	ErrUserLoginPassMismatch = NewError(1103, 401, "Login or password mismatch")

	// ErrUserBalanceNegative - общая сумма на счете не может быть отрицательной
	ErrUserBalanceNegative = NewError(1105, 402, "Insufficient funds")

	// ErrUserWithdrawnNegative - общая сумма списаний не может быть отрицательной
	ErrUserWithdrawnNegative = NewError(1106, 500, "Withdrawn amount cannot be negative")
)

// Ошибки операций 1200-1299
var (
	// ErrOperationAttrsInvalid - аттрибуты операции должны соответствовать типу операции
	ErrOperationAttrsInvalid = NewError(1200, 500, "Invalid operation attributes")

	// ErrOperationAmountInvalid - зачисления должны иметь положительные значения, а списания - отрицательные
	ErrOperationAmountInvalid = NewError(1201, 400, "Invalid operation amount sign")

	// ErrOperationUserNotExists - операция должна ссылаться на существующего пользователя
	ErrOperationUserNotExists = NewError(1202, 404, "User not exists")

	// ErrOperationOrderNumberInvalid - неверный номер заказа
	ErrOperationOrderNumberInvalid = NewError(1203, 422, "Invalid order number")

	// ErrOperationOrderNotBelongs - номер заказа уже был загружен другим пользователем
	ErrOperationOrderNotBelongs = NewError(1204, 409, "Order number belongs to another user")

	// ErrOperationOrderUsed - по заказу возможна 1 операция списания баллов и 1 операция зачисления баллов
	ErrOperationOrderUsed = NewError(1205, 409, "Order already used")

	// ErrOperationPromoUsed - пользователь может воспользоваться промо-кампанией не более 1 раза
	ErrOperationPromoUsed = NewError(1206, 409, "Promo already used")
)

// Ошибки создания промо-кампаний 1300-1399
var (
	// ErrPromoAlreadyExists - промо-кампания с таким кодом уже существует
	ErrPromoAlreadyExists = NewError(1300, 409, "Promo already exists")

	// ErrPromoRewardNotPositive - вознаграждение за промо-кампанию должно быть положительным
	ErrPromoRewardNotPositive = NewError(1301, 400, "Promo reward must be positive")

	// ErrPromoPeriodInvalid - дата начала промо-кампании должна быть меньше даты окончания промо
	ErrPromoPeriodInvalid = NewError(1302, 400, "Invalid promo period")
)

// Интеграционные ошибки 1400-1499
var (
	// ErrIntegrationTooManyRequests - слишком много запросов к внешнему сервису
	ErrIntegrationTooManyRequests = NewError(1400, 429, "Too many requests")

	// ErrIntegrationRequestFailed - ошибка при запросе к внешнему сервису
	ErrIntegrationRequestFailed = NewError(1401, 500, "Request failed")
)

// Error - ошибка приложения
type Error struct {
	error
	Code     int
	HTTPCode int
}

func NewError(code, httpCode int, message string) *Error {
	return &Error{
		error:    errors.New(message),
		Code:     code,
		HTTPCode: httpCode,
	}
}

func (e *Error) Is(tgt error) bool {
	target, ok := tgt.(*Error)
	if !ok {
		return false
	}
	return e.Code == target.Code
}
