package errors

import (
	gerrors "github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() gerrors.StackTrace
}

// New error msg
func New(message string) error {
	return gerrors.New(message)
}

// Wrap once warp stack
func Wrap(err error, message string) error {
	_, ok := err.(stackTracer)
	if ok {
		return err
	}

	return gerrors.Wrap(err, message)
}
