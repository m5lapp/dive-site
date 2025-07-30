package models

import "errors"

var (
	ErrDuplicateDiveNumber = errors.New("models: duplicate dive number for user")
	ErrDuplicateEmail      = errors.New("models: duplicate email")
	ErrInvalidCredentials  = errors.New("models: invalid credentials")
	ErrNoRecord            = errors.New("models: no matching record found")
)
