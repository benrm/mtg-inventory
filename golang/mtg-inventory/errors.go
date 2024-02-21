package inventory

import (
	"errors"
)

var (
	// ErrUserNoExist is the error returned when a user does not exist
	ErrUserNoExist = errors.New("user does not exist")

	// ErrRequestNoExist is the error returned when a request does not
	// exist
	ErrRequestNoExist = errors.New("request does not exist")
)
