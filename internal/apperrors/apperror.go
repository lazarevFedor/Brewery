// Package apperrors хранит в себе структуру ошибки приложения
package apperrors

import "net/http"

type APIError struct {
	ErrorCode string `json:"error"`
	Message   string `json:"message"`
}

type AppError struct {
	APIErr     APIError
	HTTPStatus int
	Err        error
}

func newAppError(errorCode, message string, status int, err error) *AppError {
	apiErr := APIError{
		ErrorCode: errorCode,
		Message:   message,
	}

	return &AppError{
		APIErr:     apiErr,
		HTTPStatus: status,
		Err:        err,
	}
}

func NotFound(message string, err error) *AppError {
	return newAppError(CodeNotFound, message, http.StatusNotFound, err)
}

func Validation(message string, err error) *AppError {
	return newAppError(CodeValidation, message, http.StatusUnprocessableEntity, err)
}

func Conflict(message string, err error) *AppError {
	return newAppError(CodeConflict, message, http.StatusConflict, err)
}

func Internal(err error) *AppError {
	return newAppError(CodeInternalError, "Unexpected error occurred", http.StatusInternalServerError, err)
}

func BadRequest(message string, err error) *AppError {
	return newAppError(CodeBadRequest, message, http.StatusBadRequest, err)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.APIErr.Message
}

func (e *AppError) UnWrap() error {
	return e.Err
}
