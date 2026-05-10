// Package apperrors хранит в себе структуру ошибки приложения
package apperrors

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string{
	if e.Err != nil{
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) UnWrap() error {
	return e.Err
}
