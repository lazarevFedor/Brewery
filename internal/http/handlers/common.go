package handlers

type Handlers struct {
	CategoryHandler   CategoriesHandlers
	BeersHandler      BeersHandlers
	EnumHandler       EnumHandlers
	ParametersHandler ParametersHandlers
	AggregatesHandler AggregateHandlers
	AuthHandler       AuthHandlers
}
