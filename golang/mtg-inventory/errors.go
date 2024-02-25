package inventory

import (
	"errors"
	"fmt"
)

var (
	// ErrUserNoExist is the error returned when a user does not exist
	ErrUserNoExist = errors.New("user does not exist")

	// ErrRequestNoExist is the error returned when a request does not
	// exist
	ErrRequestNoExist = errors.New("request does not exist")

	// ErrTransferNoExist is the error returned when a transfer does not
	// exist
	ErrTransferNoExist = errors.New("transfer does not exist")

	// ErrTooManyRows is returned when too many rows are submitted
	ErrTooManyRows = fmt.Errorf("more than %d rows", RowUploadLimit)

	// ErrZeroCards is returned when an uploaded row has zero or
	// fewer cards
	ErrZeroCards = errors.New("zero cards")

	// ErrUnimplemented is returned when a function is not implemented
	ErrUnimplemented = errors.New("unimplemented")
)

// RowError describes an error that takes place because of a submitted row
type RowError struct {
	Err error
	Row any
}

// Error returns a string form of the error
func (re *RowError) Error() string {
	return re.Err.Error()
}

// Unwrap returns the error this RowError wraps
func (re *RowError) Unwrap() error {
	return re.Err
}
