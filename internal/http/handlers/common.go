package handlers

const (
	BadRequest        = "BAD_REQUEST"
	InvalidJSON       = "INVALID_JSON"
	IntenalError      = "INTERNAL_ERROR"
	InvalidID         = "INVALID_ID"
	InvalidParameters = "INVALID_PARAMETERS"
)

type Handlers struct {
	CategoryHandler CategoriesHandlers
	BeersHandler    BeersHandlers
	ReviewHandler   ReviewsHandlers
	EnumHandler     EnumHandlers
}
