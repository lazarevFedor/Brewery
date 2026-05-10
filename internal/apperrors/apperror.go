// Package apperrors хранит в себе структуру ошибки приложения
package apperrors

type APIError struct {
	ErrorCode string `json:"error"`
	Message   string `json:"message"`
}

type AppError struct {
	APIErr     APIError
	HTTPStatus int
	Err        error
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
