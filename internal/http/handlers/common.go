// Package handlers содержит реализацию HTTP-обработчиков для управления пивоварней.
package handlers

// Сообщения об ошибках, которые могут возникать при обработке HTTP-запросов.
const (
	BadRequest        = "BAD_REQUEST"
	InvalidJSON       = "INVALID_JSON"
	InternalError     = "INTERNAL_ERROR"
	InvalidID         = "INVALID_ID"
	InvalidParameters = "INVALID_PARAMETERS"
)

// Handlers - структура, которая содержит все обработчики для различных маршрутов HTTP-запросов.
type Handlers struct {
	CategoryHandler   CategoriesHandlers
	BeersHandler      BeersHandlers
	ReviewHandler     ReviewsHandlers
	EnumHandler       EnumHandlers
	ParametersHandler ParametersHandlers
	AggregatesHandler AggregateHandlers
}
