// Package clierrors provides a set of error types for use in the application.
package clierrors

import "fmt"

type ExitCoder interface {
	ExitCode() int
}

type ExitError struct {
	Err  error
	Code int
}

func (e *ExitError) Error() string {
	if e == nil {
		return "unknown error"
	}
	if e.Err == nil {
		return fmt.Sprintf("unknown error (exit code %d)", e.Code)
	}

	return e.Err.Error()
}

func (e *ExitError) ExitCode() int {
	if e == nil {
		return 1
	}

	return e.Code
}

func Exit(err error, code int) error {
	if err == nil {
		return nil
	}

	return &ExitError{
		Err:  err,
		Code: code,
	}
}
