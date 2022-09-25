// Package errs - бизнес-ошибки.
package errs

import "errors"

var (
	Unauthorized        = errors.New("unauthorized")
	NotFound            = errors.New("not found")
	Validation          = errors.New("validation error")
	Duplicate           = errors.New("duplicate")
	Conflict            = errors.New("conflict")
	InsufficientBalance = errors.New("insufficient balance")
	NonExistingUser     = errors.New("non-existing user")
	NonExistingPromo    = errors.New("non-existing promo")
	Internal            = errors.New("internal error")
)
