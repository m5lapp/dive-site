package models

import (
	"errors"
	"fmt"
)

var (
	ErrUpdateConflict      = errors.New("models: conflict during update")
	ErrDuplicateDiveNumber = errors.New("models: duplicate dive number for user")
	ErrDuplicateEmail      = errors.New("models: duplicate email")
	ErrInvalidCredentials  = errors.New("models: invalid credentials")
	ErrNoRecord            = errors.New("models: no matching record found")
)

type ErrUnexpectedRowsAffected struct {
	rowsExpected int
	rowsAffected int
}

func (e *ErrUnexpectedRowsAffected) Error() string {
	msg := "models: unexpected number of rows affected, expected %d, was %d"
	return fmt.Sprintf(msg, e.rowsExpected, e.rowsAffected)
}
