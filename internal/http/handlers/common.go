package handlers

type Handlers struct {
	CategoryHandler   CategoriesHandlers
	BeersHandler      BeersHandlers
	ReviewHandler     ReviewsHandlers
	EnumHandler       EnumHandlers
	ParametersHandler ParametersHandlers
	AggregatesHandler AggregateHandlers
}
