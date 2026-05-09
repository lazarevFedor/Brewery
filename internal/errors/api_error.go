package errors

type APIError struct {
	Err     string `json:"error" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid request"`
}
