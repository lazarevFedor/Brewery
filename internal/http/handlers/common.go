package handlers

const (
	BadRequest   = "BAD_REQUEST"
	invalidJSON  = "INVALID_JSON"
	intenalError = "INTERNAL_ERROR"
	invalidID    = "INVALID_ID"
)

type Handlers struct {
	CategoryHandler CategoriesHandlers
	BeersHandler    BeersHandlers
	ReviewHandler   ReviewsHandlers
	EnumHandler     EnumHandlers
}
