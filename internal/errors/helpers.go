package errors

import "net/http"

func new(code, message string, status int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Err:        err,
	}
}

func NotFound(message string, err error) *AppError {
	return new(CodeNotFound, message, http.StatusNotFound, err)
}

func Validation(message string, err error) *AppError {
	return new(CodeValidation, message, http.StatusUnprocessableEntity, err)
}

func Conflict(message string, err error) *AppError {
	return new(CodeConflict, message, http.StatusConflict, err)
}

func Internal(err error) *AppError {
	return new(CodeInternalError, "Unexpected error occurred", http.StatusInternalServerError, err)
}

func BadRequest(message string, err error) *AppError {
	return new(CodeBadRequest, message, http.StatusBadRequest, err)
}
