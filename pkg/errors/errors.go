package errors

import (
	"errors"
	"fmt"
)

var (
	ErrValidation     = errors.New("validation error")
	ErrNotFound       = errors.New("not found")
	ErrInternal       = errors.New("internal error")
	ErrTimeout        = errors.New("timeout error")
	ErrRetryExhausted = errors.New("retry exhausted")
)

type AppError struct {
	Err     error
	Message string
	Code    string
	Details map[string]interface{}
}

func (e *AppError) Error() (msg string) {
	if e.Message != "" {
		msg = fmt.Sprintf("%s: %v", e.Message, e.Err)
		return
	}
	msg = e.Err.Error()
	return
}

func (e *AppError) Unwrap() (err error) {
	err = e.Err
	return
}

func New(err error, message string) (appErr *AppError) {
	appErr = &AppError{
		Err:     err,
		Message: message,
		Details: make(map[string]interface{}),
	}
	return
}

func (e *AppError) WithCode(code string) (appErr *AppError) {
	e.Code = code
	appErr = e
	return
}

func (e *AppError) WithDetail(key string, value interface{}) (appErr *AppError) {
	e.Details[key] = value
	appErr = e
	return
}

func IsValidation(err error) (result bool) {
	result = errors.Is(err, ErrValidation)
	return
}

func IsTimeout(err error) (result bool) {
	result = errors.Is(err, ErrTimeout)
	return
}
