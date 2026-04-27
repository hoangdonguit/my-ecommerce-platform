package errs

import "fmt"

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
	appErr, ok := err.(*AppError)
	return ok && appErr.Code == code
}

func WrapInternal(err error, message string) error {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: fmt.Sprintf("%s: %v", message, err),
	}
}
