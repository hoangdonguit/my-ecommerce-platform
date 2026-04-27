package errs

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func BadRequest(message string) error {
	return &AppError{Code: "BAD_REQUEST", Message: message}
}

func NotFound(message string) error {
	return &AppError{Code: "NOT_FOUND", Message: message}
}

func Conflict(message string) error {
	return &AppError{Code: "CONFLICT", Message: message}
}

func Internal(message string) error {
	return &AppError{Code: "INTERNAL_ERROR", Message: message}
}

func IsCode(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

func WrapInternal(err error, message string) error {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: fmt.Sprintf("%s: %v", message, err),
	}
}
