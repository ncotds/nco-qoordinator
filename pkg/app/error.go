package app

import (
	"errors"
	"fmt"
)

// Error is an error with additional data
type Error struct {
	code    string
	message string
	reason  error
}

// Err creates new instance of Error.
//
// Params:
//   - code - 'business' error code, see ErrCodeUnknown or other ErrCode* constants
//   - msg - any error description
//   - reason - wraps 'root cause' errors
func Err(code, msg string, reason ...error) Error {
	err := Error{
		code:    code,
		message: msg,
		reason:  errors.Join(reason...),
	}
	return err
}

// Code returns error-code, see ErrCode* constants
func (e Error) Code() string {
	return e.code
}

// Message returns Error's message
func (e Error) Message() string {
	return e.message
}

// Error implements std error interface
func (e Error) Error() string {
	if e.code == "" {
		return e.message
	}
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// Is checks if target's code matches current error's code
func (e Error) Is(target error) bool {
	var targetAppErr Error
	if !errors.As(target, &targetAppErr) {
		return false
	}
	isTarget := targetAppErr.code == e.code
	return isTarget
}

// Unwrap returns the reason of current error
func (e Error) Unwrap() error {
	return e.reason
}
