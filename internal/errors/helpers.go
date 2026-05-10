package apperrors

import "net/http"

func newError(code, message string, status int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Err:        err,
	}
}

func NotFound(message string, err error) *AppError {
	return newError(CodeNotFound, message, http.StatusNotFound, err)
}

func Validation(message string, err error) *AppError {
	return newError(CodeValidation, message, http.StatusUnprocessableEntity, err)
}

func Conflict(message string, err error) *AppError {
	return newError(CodeConflict, message, http.StatusConflict, err)
}

func Internal(err error) *AppError {
	return newError(CodeInternalError, "Unexpected error occurred", http.StatusInternalServerError, err)
}

func BadRequest(message string, err error) *AppError {
	return newError(CodeBadRequest, message, http.StatusBadRequest, err)
}
